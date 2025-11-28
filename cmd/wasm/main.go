//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"strconv"
	"syscall/js"

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

type ColorPickers struct {
	fg        js.Value
	bg        js.Value
	currentBG app.Color
	currentFG app.Color
}

func newColorPickers(doc js.Value, app *app.App) ColorPickers {
	bg := doc.Call("getElementById", "bgPicker")
	fg := doc.Call("getElementById", "fgPicker")
	bg.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		color := bg.Get("value").String()
		app.SetColor(0, color)
		return nil
	}))
	fg.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		color := fg.Get("value").String()
		app.SetColor(1, color)
		return nil
	}))
	return ColorPickers{bg: bg, fg: fg}
}

func (cp *ColorPickers) setColors(bg, fg app.Color) {
	if cp.currentBG != bg {
		cp.currentBG = bg
		cp.bg.Set("value", bg.ToHex())
	}

	if cp.currentFG != fg {
		cp.currentFG = fg
		cp.fg.Set("value", fg.ToHex())
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
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	len, err := wasm.app.LoadROM(buf)
	if err != nil {
		fmt.Println(err)
	}

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

func loop(this js.Value, args []js.Value) any {
	bg := wasm.app.Palette.Pixels[0]
	fg := wasm.app.Palette.Pixels[1]
	wasm.colorPickers.setColors(bg, fg)
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
