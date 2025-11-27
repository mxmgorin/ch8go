//go:build !js

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	"github.com/mxmgorin/ch8go/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

const bpp = 3

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

	app, err := NewApp(*scale)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Quit()

	fmt.Println("ch8go SDL2")
	fmt.Printf("ROM: %s\n", *romPath)

	app.Run(rom)
}

type App struct {
	window   *sdl.Window
	texture  *sdl.Texture
	renderer *sdl.Renderer
	emu      *chip8.VM
	rgbBuf   []byte
	scale    int
}

func NewApp(scale int) (*App, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		return nil, err
	}
	emu := chip8.NewVM()
	window, err := sdl.CreateWindow("ch8go",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(emu.Display.Width*scale),
		int32(emu.Display.Height*scale),
		sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_STREAMING, int32(emu.Display.Width), int32(emu.Display.Height))
	if err != nil {
		return nil, err
	}
	bufSize := emu.Display.Width * emu.Display.Height * bpp
	return &App{window: window, renderer: renderer, texture: texture, emu: emu, rgbBuf: make([]byte, bufSize)}, nil
}

func (a *App) Quit() {
	a.texture.Destroy()
	a.renderer.Destroy()
	a.window.Destroy()
	sdl.Quit()
}

func (a *App) Run(rom []byte) error {
	if err := a.emu.LoadRom(rom); err != nil {
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
					handleKey(ev.Keysym.Sym, &a.emu.Keypad, true)
				case sdl.KEYUP:
					handleKey(ev.Keysym.Sym, &a.emu.Keypad, false)
				}
			}
		}

		if a.emu.RunFrame() {
			a.draw()
		}

		elapsed := time.Since(frameStart)
		if elapsed < frameDelay {
			time.Sleep(frameDelay - elapsed)
		}
	}

	return nil
}

func (a *App) draw() {
	for i, px := range a.emu.Display.Pixels {
		v := byte(0)
		if px != 0 {
			v = 255
		}

		a.rgbBuf[i*bpp+0] = v
		a.rgbBuf[i*bpp+1] = v
		a.rgbBuf[i*bpp+2] = v
	}

	a.texture.Update(nil, unsafe.Pointer(&a.rgbBuf[0]), a.emu.Display.Width*bpp)
	a.renderer.Clear()
	a.renderer.Copy(a.texture, nil, nil)
	a.renderer.Present()
}

func handleKey(key sdl.Keycode, keypad *chip8.Keypad, down bool) {
	if k, ok := keymap[key]; ok {
		keypad.HandleKey(k, down)
	}
}
