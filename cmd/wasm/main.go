//go:build js && wasm

package main

import (
	"fmt"
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
	ctx       js.Value
	imageData js.Value
}

func (p *CanvasPainter) Init(w, h int) error {
	scale := 5
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "chip8-canvas")
	canvas.Set("width", w)
	canvas.Set("height", h)
	canvasStyle := canvas.Get("style")
	canvasStyle.Set("transform", fmt.Sprintf("scale(%d)", scale))

	screen := doc.Call("getElementById", "chip8-screen")
	screen.Set("width", w*scale)
	screen.Set("height", h*scale)

	p.ctx = canvas.Call("getContext", "2d")
	p.imageData = p.ctx.Call("createImageData", w, h)
	return nil
}

func (p *CanvasPainter) Destroy() {}

func (p *CanvasPainter) Paint(rgbaBuf []byte, w, h int) {
	js.CopyBytesToJS(p.imageData.Get("data"), rgbaBuf)
	p.ctx.Call("putImageData", p.imageData, 0, 0)
}

type WASM struct {
	loopFunc js.Func
	app      *app.App
}

func newWASM() WASM {
	// Export ROM loader
	js.Global().Set("chip8_loadROM", js.FuncOf(loadROM))

	// Handle keyboard
	win := js.Global().Get("window")
	win.Call("addEventListener", "keydown", js.FuncOf(onKeyDown))
	win.Call("addEventListener", "keyup", js.FuncOf(onKeyUp))

	// Animation loop (must persist function or GC will kill it)
	loopFunc := js.FuncOf(loop)
	js.Global().Call("requestAnimationFrame", loopFunc)

	painter := &CanvasPainter{}
	app, _ := app.NewApp(painter)

	return WASM{
		app:      app,
		loopFunc: loopFunc,
	}
}

func (wasm *WASM) run() {
	// Keep WASM alive
	select {}
}

func loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	wasm.app.LoadROM(buf)
	fmt.Println("ROM loaded")
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
	if !wasm.app.HasROM() {
		// Don't run CPU until ROM exists
		js.Global().Call("requestAnimationFrame", wasm.loopFunc)
		return nil
	}

	if wasm.app.VM.RunFrame() {
		wasm.app.Paint()
	}

	// Schedule next frame
	js.Global().Call("requestAnimationFrame", wasm.loopFunc)
	return nil
}

func main() {
	fmt.Println("ch8go WASM")

	wasm = newWASM()
	wasm.run()
}
