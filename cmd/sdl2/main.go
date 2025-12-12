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

var keymap = map[sdl.Keycode]chip8.Key{
	sdl.K_1: chip8.Key1,
	sdl.K_2: chip8.Key2,
	sdl.K_3: chip8.Key3,
	sdl.K_4: chip8.KeyC,

	sdl.K_q: chip8.Key4,
	sdl.K_w: chip8.Key5,
	sdl.K_e: chip8.Key6,
	sdl.K_r: chip8.KeyD,

	sdl.K_a: chip8.Key7,
	sdl.K_s: chip8.Key8,
	sdl.K_d: chip8.Key9,
	sdl.K_f: chip8.KeyE,

	sdl.K_z: chip8.KeyA,
	sdl.K_x: chip8.Key0,
	sdl.K_c: chip8.KeyB,
	sdl.K_v: chip8.KeyF,
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
