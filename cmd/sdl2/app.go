package main

import (
	"time"

	"github.com/mxmgorin/ch8go/pkg/host"
	"github.com/veandco/go-sdl2/sdl"
)

type App struct {
	*host.Emu
	painter *Painter
}

func newApp() (*App, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	emu, err := host.NewEmu()
	if err != nil {
		return nil, err
	}

	size := emu.VM.Display.Size()
	painter, err := newPainter(size.Width, size.Height, 10)
	if err != nil {
		return nil, err
	}

	return &App{Emu: emu, painter: painter}, nil
}

func (a *App) Quit() {
	a.painter.Destroy()
	sdl.Quit()
}

func (a *App) Run() error {
	frameDelay := time.Second / 60 // target 60 FPS

	running := true
	for running {
		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				switch ev.Type {
				case sdl.KEYDOWN:
					handleKey(ev.Keysym.Sym, &a.VM.Keypad, true)
				case sdl.KEYUP:
					handleKey(ev.Keysym.Sym, &a.VM.Keypad, false)
				}
			}
		}

		fb := a.RunFrame()
		a.painter.Paint(fb)

		elapsed := time.Since(frameStart)
		if elapsed < frameDelay {
			time.Sleep(frameDelay - elapsed)
		}
	}

	return nil
}
