//go:build !js

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

type Sdl2App struct {
	app *app.App
}

func NewSdl2App(scale int) (*Sdl2App, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	app, err := app.NewApp(&Sdl2Painter{scale: 10})
	if err != nil {
		return nil, err
	}

	return &Sdl2App{app: app}, nil
}

func (a *Sdl2App) Quit() {
	a.app.Quit()
	sdl.Quit()
}

func (a *Sdl2App) Run(rom []byte) error {
	if _, err := a.app.LoadROM(rom); err != nil {
		return err
	}
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
					handleKey(ev.Keysym.Sym, &a.app.VM.Keypad, true)
				case sdl.KEYUP:
					handleKey(ev.Keysym.Sym, &a.app.VM.Keypad, false)
				}
			}
		}

		a.app.PaintFrame()

		elapsed := time.Since(frameStart)
		if elapsed < frameDelay {
			time.Sleep(frameDelay - elapsed)
		}
	}

	return nil
}

type Sdl2Painter struct {
	window   *sdl.Window
	texture  *sdl.Texture
	renderer *sdl.Renderer
	scale    int
}

func (p *Sdl2Painter) Init(width, height int) error {
	window, err := sdl.CreateWindow("ch8go SDL2",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(width*p.scale),
		int32(height*p.scale),
		sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	p.window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	p.renderer = renderer

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		int32(width),
		int32(height))
	if err != nil {
		return err
	}
	p.texture = texture

	return nil
}

func (p *Sdl2Painter) Paint(rgbaBuf []byte, width, height int) {
	pitch := width * 4
	p.texture.Update(nil, unsafe.Pointer(&rgbaBuf[0]), pitch)
	p.renderer.Clear()
	p.renderer.Copy(p.texture, nil, nil)
	p.renderer.Present()
}

func (p *Sdl2Painter) Destroy() {
	p.texture.Destroy()
	p.renderer.Destroy()
	p.window.Destroy()
}

func handleKey(key sdl.Keycode, keypad *chip8.Keypad, down bool) {
	if k, ok := keymap[key]; ok {
		keypad.HandleKey(k, down)
	}
}

func main() {
	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	scale := flag.Int("scale", 12, "window scale")
	flag.Parse()

	if *romPath == "" {
		log.Fatal("You must provide a ROM: --rom path/to/file.ch8")
	}

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		log.Fatal(err)
	}

	app, err := NewSdl2App(*scale)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Quit()

	fmt.Println("ch8go SDL2")
	fmt.Printf("ROM: %s\n", *romPath)

	if err := app.Run(rom); err != nil {
		log.Fatal(err)
	}
}
