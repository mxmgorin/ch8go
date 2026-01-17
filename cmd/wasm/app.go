//go:build js && wasm

package main

import (
	"log"
	"log/slog"
	"path/filepath"
	"syscall/js"

	"github.com/mxmgorin/ch8go/pkg/host"
)

type App struct {
	emu               *host.Emu
	painter           Painter
	audio             Audio
	input             Input
	confOverlay       ConfOverlay
	palettePicker     PalettePicker
	runFrameFunc      js.Func
	togglePauseIconEl js.Value
	pauseOverlayEl    js.Value
	keyChan           chan KeyEvent
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

	jsGlobal := js.Global()
	jsGlobal.Set("fillROMs", js.FuncOf(populateROMs))
	doc := js.Global().Get("document")
	win := js.Global().Get("window")
	keyChan := make(chan KeyEvent, 32)

	a := App{
		palettePicker:     newPalettePicker(doc, &emu.Palette),
		painter:           painter,
		audio:             newAudio(jsGlobal, &emu.VM.Audio),
		input:             newInput(win, keyChan),
		confOverlay:       newConfOverlay(doc, emu.VM),
		togglePauseIconEl: doc.Call("getElementById", "toggle-pause-icon"),
		pauseOverlayEl:    doc.Call("getElementById", "pause-overlay"),
		keyChan:           keyChan,
		emu:               emu,
	}

	jsGlobal.Set("chip8_loadROM", js.FuncOf(a.loadROM))
	togglePauseBtn := doc.Call("getElementById", "toggle-pause-btn")
	togglePauseBtn.Call("addEventListener", "click", js.FuncOf(a.togglePause))

	// Animation loop (must persist function or GC will kill it)
	a.runFrameFunc = js.FuncOf(a.runFrame)

	return a
}

func (a *App) loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	name := args[1].String()
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	_, err := a.emu.LoadROM(buf, filepath.Ext(name))
	if err != nil {
		slog.Error("Failed to LoadROM", "err", err)
	}

	a.palettePicker.setColors(&a.emu.Palette.Pixels)
	a.confOverlay.setTickrate(a.emu.VM.Tickrate())
	a.confOverlay.setQuirks(a.emu.VM.CPU.Quirks)
	setROMInfo(a.emu.ROMInfo())

	return nil
}

func (a *App) togglePause(this js.Value, args []js.Value) any {
	a.emu.Paused = !a.emu.Paused

	if a.emu.Paused {
		a.pauseOverlayEl.Get("classList").Call("add", "active")
		a.togglePauseIconEl.Set("src", "./icons/play-icon.svg")
	} else {
		a.pauseOverlayEl.Get("classList").Call("remove", "active")
		a.togglePauseIconEl.Set("src", "./icons/pause-icon.svg")
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
