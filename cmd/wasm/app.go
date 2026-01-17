//go:build js && wasm

package main

import (
	"log"
	"syscall/js"

	"github.com/mxmgorin/ch8go/pkg/host"
)

type App struct {
	emu             *host.Emu
	painter         Painter
	audio           Audio
	input           Input
	runFrameFunc    js.Func
	pauseOverlay    js.Value
	confOverlay     ConfOverlay
	palettePicker   PalettePicker
	togglePauseIcon js.Value
	keyChan         chan KeyEvent
}

func newApp() App {
	emu, err := host.NewEmu()
	if err != nil {
		log.Fatal(err)
	}

	displaySize := emu.VM.Display.Size()
	painter, err := newPainter(displaySize.Width, displaySize.Height)
	if err != nil {
		log.Fatal(err)
	}

	// Export ROM loader
	js.Global().Set("chip8_loadROM", js.FuncOf(loadROM))
	// ROMs select
	js.Global().Set("fillROMs", js.FuncOf(populateROMs))

	doc := js.Global().Get("document")
	palettePicker := newPalettePicker(doc, &emu.Palette)
	confOverlay := newConfOverlay(doc, emu.VM)

	// Pause, resume
	pauseOverlay := js.Global().Get("document").Call("getElementById", "pause-overlay")
	togglePauseIcon := doc.Call("getElementById", "toggle-pause-icon")

	a := App{
		emu:             emu,
		palettePicker:   palettePicker,
		painter:         painter,
		audio:           newAudio(),
		input:           newInput(),
		confOverlay:     confOverlay,
		togglePauseIcon: togglePauseIcon,
		pauseOverlay:    pauseOverlay,
		keyChan:         make(chan KeyEvent, 32),
	}

	togglePauseBtn := doc.Call("getElementById", "toggle-pause-btn")
	togglePauseBtn.Call("addEventListener", "click", js.FuncOf(a.togglePause))

	// Animation loop (must persist function or GC will kill it)
	a.runFrameFunc = js.FuncOf(a.runFrame)

	return a
}

func (a *App) togglePause(this js.Value, args []js.Value) any {
	a.emu.Paused = !a.emu.Paused

	if a.emu.Paused {
		a.pauseOverlay.Get("classList").Call("add", "active")
		a.togglePauseIcon.Set("src", "./icons/play-icon.svg")
	} else {
		a.pauseOverlay.Get("classList").Call("remove", "active")
		a.togglePauseIcon.Set("src", "./icons/pause-icon.svg")
	}

	return nil
}

func (a *App) handleKey(evt KeyEvent) {
	if evt.Pressed {
		a.emu.VM.Keypad.Press(evt.Key)
	} else {
		a.emu.VM.Keypad.Release(evt.Key)
	}
}

func (a *App) runFrame(this js.Value, args []js.Value) any {
	drainChan(a.keyChan, a.handleKey)
	fb := a.emu.RunFrame()
	a.painter.Paint(fb)
	// Schedule next frame
	js.Global().Call("requestAnimationFrame", a.runFrameFunc)
	return nil
}

// Run main loop
func (a *App) run() {
	js.Global().Call("requestAnimationFrame", a.runFrameFunc)
	// Keep WASM alive
	select {}
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
