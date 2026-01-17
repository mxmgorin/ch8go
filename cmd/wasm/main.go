//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"syscall/js"
	"unsafe"

	"github.com/mxmgorin/ch8go/pkg/host"
)

var (
	audioBuf = make([]float32, 0)
	app      App
)

type App struct {
	runFrameFunc    js.Func
	emu             *host.Emu
	colorPickers    ColorPickers
	painter         Painter
	ConfOverlay     ConfOverlay
	togglePauseIcon js.Value
	pauseOverlay    js.Value
	keyChan         chan KeyEvent
}

func newApp() App {
	emu, err := host.NewEmu()
	if err != nil {
		log.Fatal(err)
	}

	size := emu.VM.Display.Size()
	painter, err := newPainter(size.Width, size.Height)
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
	colorPickers := newColorPickers(doc, &emu.Palette)

	// Conf
	conf := newConf(doc, emu.VM)

	// Audio
	js.Global().Set("fillAudio", js.FuncOf(fillAudio))
	js.Global().Set("startAudio", js.FuncOf(startAudio))

	// Pause, resume
	pauseOverlay := js.Global().Get("document").Call("getElementById", "pause-overlay")
	togglePauseIcon := doc.Call("getElementById", "toggle-pause-icon")
	togglePauseBtn := doc.Call("getElementById", "toggle-pause-btn")
	togglePauseBtn.Call("addEventListener", "click", js.FuncOf(togglePause))

	// Roms
	js.Global().Set("fillROMs", js.FuncOf(fillROMs))

	// Animation loop (must persist function or GC will kill it)
	runFrameFunc := js.FuncOf(runFrame)

	return App{
		emu:             emu,
		runFrameFunc:    runFrameFunc,
		colorPickers:    colorPickers,
		painter:         painter,
		ConfOverlay:     conf,
		togglePauseIcon: togglePauseIcon,
		pauseOverlay:    pauseOverlay,
		keyChan:         make(chan KeyEvent, 32),
	}
}

// Run main loop
func (a *App) run() {
	js.Global().Call("requestAnimationFrame", a.runFrameFunc)
	// Keep WASM alive
	select {}
}

// Set ROM info in overlay
func setROMInfo() {
	text := app.emu.ROMInfo()
	doc := js.Global().Get("document")
	info := doc.Call("getElementById", "info-overlay")
	info.Set("innerHTML", text)
}

func loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	name := args[1].String()
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	_, err := app.emu.LoadROM(buf, filepath.Ext(name))
	if err != nil {
		fmt.Println(err)
	}

	app.colorPickers.setColors(&app.emu.Palette.Pixels)
	app.ConfOverlay.setTickrate(app.emu.VM.Tickrate())
	app.ConfOverlay.setQuirks(app.emu.VM.CPU.Quirks)
	setROMInfo()

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
	app.emu.VM.Audio.Output(audioBuf, freq)

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

func runFrame(this js.Value, args []js.Value) any {
	drainChan(app.keyChan, handleKey)
	fb := app.emu.RunFrame()
	app.painter.Paint(fb)
	// Schedule next frame
	js.Global().Call("requestAnimationFrame", app.runFrameFunc)
	return nil
}

func fillROMs(this js.Value, args []js.Value) any {
	selectEl := js.Global().Get("document").Call("getElementById", "roms")
	selectEl.Set("innerHTML", "")

	for _, r := range Roms {
		opt := js.Global().Get("document").Call("createElement", "option")
		opt.Set("value", r.Path)
		opt.Set("textContent", r.Name)
		selectEl.Call("appendChild", opt)
	}

	return nil
}

func togglePause(this js.Value, args []js.Value) any {
	app.emu.Paused = !app.emu.Paused

	if app.emu.Paused {
		app.pauseOverlay.Get("classList").Call("add", "active")
		app.togglePauseIcon.Set("src", "./icons/play-icon.svg")
	} else {
		app.pauseOverlay.Get("classList").Call("remove", "active")
		app.togglePauseIcon.Set("src", "./icons/pause-icon.svg")
	}

	return nil
}

func main() {
	slog.Info("ch8go WASM")
	app = newApp()
	app.run()
}
