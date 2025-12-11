//go:build !js

package main

import (
	"log"
	"log/slog"
	"time"
	"unsafe"

	"github.com/mxmgorin/ch8go/app"
	"github.com/mxmgorin/ch8go/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

var keymap = map[sdl.Keycode]byte{
	sdl.K_1: 0x1,
	sdl.K_2: 0x2,
	sdl.K_3: 0x3,
	sdl.K_4: 0xC,

	sdl.K_q: 0x4,
	sdl.K_w: 0x5,
	sdl.K_e: 0x6,
	sdl.K_r: 0xD,

	sdl.K_a: 0x7,
	sdl.K_s: 0x8,
	sdl.K_d: 0x9,
	sdl.K_f: 0xE,

	sdl.K_z: 0xA,
	sdl.K_x: 0x0,
	sdl.K_c: 0xB,
	sdl.K_v: 0xF,
}

type Sdl2Painter struct {
	window   *sdl.Window
	texture  *sdl.Texture
	renderer *sdl.Renderer
	scale    int
}

func newPainter(width, height, scale int) (*Sdl2Painter, error) {
	window, err := sdl.CreateWindow("ch8go SDL2",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(width*scale),
		int32(height*scale),
		sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	p := Sdl2Painter{}
	p.window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}
	p.renderer = renderer

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		int32(width),
		int32(height))
	if err != nil {
		return nil, err
	}
	p.texture = texture

	return &p, nil
}

func (p *Sdl2Painter) Paint(fb *app.FrameBuffer) {
	p.texture.Update(nil, unsafe.Pointer(&fb.Pixels[0]), fb.Pitch())
	p.renderer.Clear()
	p.renderer.Copy(p.texture, nil, nil)
	p.renderer.Present()
}

func (p *Sdl2Painter) Destroy() {
	p.texture.Destroy()
	p.renderer.Destroy()
	p.window.Destroy()
}

type Sdl2App struct {
	*app.App
	painter *Sdl2Painter
}

func NewSdl2App(scale int) (*Sdl2App, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	app, err := app.NewApp()
	if err != nil {
		return nil, err
	}

	painter, err := newPainter(app.VM.Display.Width, app.VM.Display.Height, 10)
	if err != nil {
		return nil, err
	}

	return &Sdl2App{App: app, painter: painter}, nil
}

func (a *Sdl2App) Quit() {
	a.painter.Destroy()
	sdl.Quit()
}

func (a *Sdl2App) Run() error {
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

func handleKey(key sdl.Keycode, keypad *chip8.Keypad, down bool) {
	if k, ok := keymap[key]; ok {
		keypad.HandleKey(k, down)
	}
}

func main() {
	slog.Info("ch8go SDL2")

	romPath, scale := app.ParseFlags()
	a, err := NewSdl2App(scale)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Quit()

	if _, err := a.ReadROM(romPath); err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
