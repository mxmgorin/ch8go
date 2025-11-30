//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"strconv"
	"syscall/js"
	"unsafe"

	"github.com/mxmgorin/ch8go/app"
)

var (
	wasm   WASM
	keymap = map[string]byte{
		"1": 0x1, "2": 0x2, "3": 0x3, "4": 0xC,
		"q": 0x4, "w": 0x5, "e": 0x6, "r": 0xD,
		"a": 0x7, "s": 0x8, "d": 0x9, "f": 0xE,
		"z": 0xA, "x": 0x0, "c": 0xB, "v": 0xF,
	}
)

type CanvasPainter struct {
	ctx           js.Value
	imageData     js.Value
	screen        js.Value
	canvas        js.Value
	screenBgColor *app.Color
	width         int
	height        int
}

func (p *CanvasPainter) setScreenBg(color app.Color) {
	if p.screenBgColor == nil || *p.screenBgColor != color {
		hex := color.ToHex()
		p.screen.Get("style").Set("background", hex)
		p.screenBgColor = &color
	}
}

func (p *CanvasPainter) Init(w, h int) error {
	p.width = w
	p.height = h
	doc := js.Global().Get("document")

	p.canvas = doc.Call("getElementById", "chip8-canvas")
	p.canvas.Set("width", w)
	p.canvas.Set("height", h)

	p.screen = doc.Call("getElementById", "chip8-screen")

	scaleInput := doc.Call("getElementById", "scaleInput")
	value := scaleInput.Get("value").String()
	p.setScale(value)
	scaleInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := scaleInput.Get("value").String()
		p.setScale(value)
		return nil
	}))

	p.ctx = p.canvas.Call("getContext", "2d")
	p.imageData = p.ctx.Call("createImageData", w, h)
	return nil
}

func (p *CanvasPainter) setScale(value string) {
	scale, err := strconv.Atoi(value)
	if err != nil {
		fmt.Println("Invalid scale:", value)
	}

	canvasStyle := p.canvas.Get("style")
	canvasStyle.Set("transform", fmt.Sprintf("scale(%d)", scale))

	screenStyle := p.screen.Get("style")
	screenStyle.Set("width", fmt.Sprintf("%dpx", p.width*scale))
	screenStyle.Set("height", fmt.Sprintf("%dpx", p.height*scale))
}

func (p *CanvasPainter) Destroy() {}

func (p *CanvasPainter) Paint(fb *app.FrameBuffer) {
	p.setScreenBg(fb.SoundColor)
	js.CopyBytesToJS(p.imageData.Get("data"), fb.Pixels)
	p.ctx.Call("putImageData", p.imageData, 0, 0)
}

type ColorPickers [4]js.Value

func newColorPickers(doc js.Value, app *app.App) ColorPickers {
	pickers := ColorPickers{}
	pickers[0] = doc.Call("getElementById", "bgPicker")
	pickers[1] = doc.Call("getElementById", "fgPicker")
	pickers[2] = doc.Call("getElementById", "c3Picker")
	pickers[3] = doc.Call("getElementById", "c4Picker")

	pickers[0].Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		color := pickers[0].Get("value").String()
		app.SetColor(0, color)
		return nil
	}))
	pickers[1].Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		color := pickers[1].Get("value").String()
		app.SetColor(1, color)
		return nil
	}))
	pickers[2].Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		color := pickers[2].Get("value").String()
		app.SetColor(2, color)
		return nil
	}))
	pickers[3].Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		color := pickers[3].Get("value").String()
		app.SetColor(3, color)
		return nil
	}))

	return pickers
}

func (cp *ColorPickers) setColors(colors [4]app.Color) {
	for i := range colors {
		cp[i].Set("value", colors[i].ToHex())
	}
}

type WASM struct {
	loopFunc     js.Func
	app          *app.App
	colorPickers ColorPickers
}

func newWASM() WASM {
	app, err := app.NewApp(&CanvasPainter{})

	if err != nil {
		log.Fatal(err)
	}

	// Export ROM loader
	js.Global().Set("chip8_loadROM", js.FuncOf(loadROM))

	// Handle keyboard
	win := js.Global().Get("window")
	win.Call("addEventListener", "keydown", js.FuncOf(onKeyDown))
	win.Call("addEventListener", "keyup", js.FuncOf(onKeyUp))

	// Palette
	doc := js.Global().Get("document")
	colorPickers := newColorPickers(doc, app)

	js.Global().Set("fillAudio", js.FuncOf(fillAudio))

	// Animation loop (must persist function or GC will kill it)
	loopFunc := js.FuncOf(loop)

	return WASM{
		app:          app,
		loopFunc:     loopFunc,
		colorPickers: colorPickers,
	}
}

func (wasm *WASM) run() {
	js.Global().Call("requestAnimationFrame", wasm.loopFunc)
	// Keep WASM alive
	select {}
}

func loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	name := args[1].String()
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	len, err := wasm.app.LoadROM(buf, app.PlaformFromPath(name))
	if err != nil {
		fmt.Println(err)
	}

	wasm.colorPickers.setColors(wasm.app.Palette.Pixels)

	fmt.Println("ROM loaded:", len, "bytes")
	return nil
}

func onKeyDown(this js.Value, args []js.Value) any {
	key := args[0].Get("key").String()
	if k, ok := keymap[key]; ok {
		wasm.app.VM.Keypad.Press(k)
		args[0].Call("preventDefault")
	}
	return nil
}

func onKeyUp(this js.Value, args []js.Value) any {
	key := args[0].Get("key").String()
	if k, ok := keymap[key]; ok {
		wasm.app.VM.Keypad.Release(k)
		args[0].Call("preventDefault")
	}
	return nil
}

func fillAudio(this js.Value, args []js.Value) any {
	n := args[0].Int()

	// produce float32 audio
	buf := make([]float32, n)
	wasm.app.VM.OutputAudio(buf, 44100.0)

	// Create Float32Array
	float32Array := js.Global().Get("Float32Array").New(n)

	// Get the underlying Uint8Array buffer
	byteArray := js.Global().Get("Uint8Array").New(float32Array.Get("buffer"))

	// Copy float32 bytes → JS buffer
	src := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), n*4)
	js.CopyBytesToJS(byteArray, src)

	return float32Array
}

func loop(this js.Value, args []js.Value) any {
	wasm.app.PaintFrame()

	// Schedule next frame
	js.Global().Call("requestAnimationFrame", wasm.loopFunc)
	return nil
}

func main() {
	fmt.Println("ch8go WASM")
	wasm = newWASM()
	wasm.run()
}
