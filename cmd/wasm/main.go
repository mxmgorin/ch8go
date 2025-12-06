//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strconv"
	"syscall/js"
	"unsafe"

	"github.com/mxmgorin/ch8go/app"
)

var (
	audioBuf = make([]float32, 0)
	wasm     WASM
	keymap   = map[string]byte{
		"1": 0x1, "2": 0x2, "3": 0x3, "4": 0xC,
		"q": 0x4, "w": 0x5, "e": 0x6, "r": 0xD,
		"a": 0x7, "s": 0x8, "d": 0x9, "f": 0xE,
		"z": 0xA, "x": 0x0, "c": 0xB, "v": 0xF,
		"ArrowUp": 0x5, "ArrowDown": 0x8, "ArrowLeft": 0x7, "ArrowRight": 0x9,
		" ": 0x6,
	}
	roms = []struct {
		Path string
		Name string
	}{
		{"roms/xo/skyward.ch8", "Skyward"},
		{"roms/xo/superneatboy.ch8", "Super Neat Boy"},
		{"roms/xo/garlicscape.ch8", "Garlic Scape"},
		{"roms/xo/octoma.ch8", "Octoma"},
		{"roms/xo/t8nks.ch8", "T8nks"},
		{"roms/xo/octopeg.ch8", "Octopeg"},
		{"roms/ch/danm8ku.ch8", "Danm8ku"},
		{"roms/ch/octogon.ch8", "Octogon"},
		{"roms/ch/supersquare.ch8", "Super Square"},
		{"roms/ch/down8.ch8", "Down8"},
		{"roms/ch/slipperyslope.ch8", "Slippery Slope"},
		{"roms/ch/rockto.ch8", "Rockto"},
		{"roms/ch/sub8.ch8", "Sub8"},
		{"roms/ch/DVN8.ch8", "DVN8"},
		{"roms/ch/flightrunner.ch8", "Flight Runner"},
		{"roms/ch/glitchGhost.ch8", "Glitch Ghost"},
		{"roms/ch/turnover77.ch8", "Turn Over 77"},
		{"roms/ch/blackrainbow.ch8", "Black Rainbow"},
		{"roms/ch/binding.ch8", "Binding"},
		{"roms/ch/br8kout.ch8", "Br8kout"},
		{"roms/ch/spacejam.ch8", "Space Jam"},
		{"roms/ch/octovore.ch8", "Octovore"},
		{"roms/ch/INVADERS", "Invaders"},
		{"roms/ch/TETRIS", "Tetris"},
		{"roms/ch/snake.ch8", "Snake"},
		{"roms/ch/TANK", "Tank"},
		{"roms/xo/D8GN.ch8", "D8GN"},
		{"roms/xo/civiliz8n.ch8", "Civiliz8n"},
		{"roms/xo/clostro.ch8", "Clostro"},
		{"roms/xo/sneaksurround.ch8", "Sneak Surround"},
		{"roms/xo/chickenScratch.ch8", "Chicken Scratch"},
		{"roms/sc/ANT", "Ant"},
		{"roms/sc/sweetcopter.ch8", "Sw8Copter"},
		{"roms/xo/tapeworm.ch8", "Tapeworm"},
		{"roms/xo/snake.ch8", "xSnake"},
		{"roms/xo/alien-inv8sion.ch8", "Alien Inv8sion (Timendus)"},
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

func newPainter(w, h int) (CanvasPainter, error) {
	p := CanvasPainter{}
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
	return p, nil
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

func (cp *ColorPickers) setColors(colors *[16]app.Color) {
	for i := range cp {
		cp[i].Set("value", colors[i].ToHex())
	}
}

type KeyEvent struct {
	Key     byte
	Pressed bool
}

type WASM struct {
	frameFunc    js.Func
	app          *app.App
	colorPickers ColorPickers
	painter      CanvasPainter
	KeyChan      chan KeyEvent
}

func newWASM() WASM {
	app, err := app.NewApp()

	if err != nil {
		log.Fatal(err)
	}

	painter, err := newPainter(app.VM.Display.Width, app.VM.Display.Height)

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

	// Audio
	js.Global().Set("fillAudio", js.FuncOf(fillAudio))
	js.Global().Set("startAudio", js.FuncOf(startAudio))

	// Roms
	js.Global().Set("fillROMs", js.FuncOf(fillROMs))

	// Animation loop (must persist function or GC will kill it)
	frameFunc := js.FuncOf(frame)

	return WASM{
		app:          app,
		frameFunc:    frameFunc,
		colorPickers: colorPickers,
		painter:      painter,
		KeyChan:      make(chan KeyEvent, 32),
	}
}

func (wasm *WASM) run() {
	js.Global().Call("requestAnimationFrame", wasm.frameFunc)
	// Keep WASM alive
	select {}
}

func loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	name := args[1].String()
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	_, err := wasm.app.LoadROM(buf, filepath.Ext(name))
	if err != nil {
		fmt.Println(err)
	}

	wasm.colorPickers.setColors(&wasm.app.Palette.Pixels)

	return nil
}

func onKeyDown(this js.Value, args []js.Value) any {
	key := args[0].Get("key").String()
	if k, ok := keymap[key]; ok {
		wasm.KeyChan <- KeyEvent{Key: k, Pressed: true}
		args[0].Call("preventDefault")
	}
	return nil
}

func onKeyUp(this js.Value, args []js.Value) any {
	key := args[0].Get("key").String()
	if k, ok := keymap[key]; ok {
		wasm.KeyChan <- KeyEvent{Key: k, Pressed: false}
		args[0].Call("preventDefault")
	}
	return nil
}

func startAudio(this js.Value, args []js.Value) any {
	size := args[0].Int()
	audioBuf = make([]float32, size)
	return nil
}

func fillAudio(this js.Value, args []js.Value) any {
	out := args[0] // JS Float32Array
	freq := args[1].Float()
	wasm.app.VM.Audio.Output(audioBuf, freq)

	outBuffer := js.Global().Get("Uint8Array").New(
		out.Get("buffer"),
		out.Get("byteOffset"),
		out.Get("byteLength"),
	)
	bufPointer := unsafe.Pointer(&audioBuf[0])
	byteBuf := unsafe.Slice((*byte)(bufPointer), len(audioBuf)*4)
	js.CopyBytesToJS(outBuffer, byteBuf)

	return nil
}

func frame(this js.Value, args []js.Value) any {
	drainChan(wasm.KeyChan, handleKey)
	fb := wasm.app.RunFrame()
	wasm.painter.Paint(fb)
	// Schedule next frame
	js.Global().Call("requestAnimationFrame", wasm.frameFunc)
	return nil
}

func handleKey(evt KeyEvent) {
	if evt.Pressed {
		wasm.app.VM.Keypad.Press(evt.Key)
	} else {
		wasm.app.VM.Keypad.Release(evt.Key)
	}
}
func fillROMs(this js.Value, args []js.Value) any {
	selectEl := js.Global().Get("document").Call("getElementById", "roms")
	selectEl.Set("innerHTML", "")

	for _, r := range roms {
		opt := js.Global().Get("document").Call("createElement", "option")
		opt.Set("value", r.Path)
		opt.Set("textContent", r.Name)
		selectEl.Call("appendChild", opt)
	}

	return nil
}

func drainChan[T any](ch <-chan T, fn func(T)) {
	for {
		select {
		case v := <-ch:
			fn(v)
		default:
			return
		}
	}
}

func main() {
	slog.Info("ch8go WASM")
	wasm = newWASM()
	wasm.run()
}
