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

const (
	WindowScale = 12
)

func main() {
	fmt.Println("Running sdl2")
	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	hz := flag.Int("hz", 500, "cpu cycles per second")
	flag.Parse()

	if *romPath == "" {
		log.Fatal("You must provide a ROM: --rom path/to/file.ch8")
	}

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("CHIP-8 SDL2")
	fmt.Printf("ROM: %s\n", *romPath)
	fmt.Printf("Speed: %d Hz\n", *hz)

	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Quit()
	app.Run(rom, *hz)
}

type App struct {
	Window   *sdl.Window
	Texture  *sdl.Texture
	Renderer *sdl.Renderer
	Emu      *chip8.Emu
	rgbBuf   []byte
}

func NewApp() (*App, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		return nil, err
	}
	emu := chip8.NewEmu()
	window, err := sdl.CreateWindow("CHIP-8",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(emu.Display.Width*WindowScale),
		int32(emu.Display.Height*WindowScale),
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
	bufSize := emu.Display.Width * emu.Display.Height * 3
	return &App{Window: window, Renderer: renderer, Texture: texture, Emu: emu, rgbBuf: make([]byte, bufSize)}, nil
}

func (a *App) Quit() {
	a.Texture.Destroy()
	a.Renderer.Destroy()
	a.Window.Destroy()
	sdl.Quit()
}

func (a *App) Run(rom []byte, hz int) error {
	if err := a.Emu.LoadRom(rom); err != nil {
		return err
	}
	cycleDelay := time.Second / time.Duration(hz)

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
					handleKey(ev.Keysym.Sym, &a.Emu.Keypad, true)
				case sdl.KEYUP:
					handleKey(ev.Keysym.Sym, &a.Emu.Keypad, false)
				}
			}
		}

		a.Emu.Step()
		a.draw()

		elapsed := time.Since(frameStart)
		if elapsed < cycleDelay {
			time.Sleep(cycleDelay - elapsed)
		}
	}

	return nil
}

func (a *App) draw() {
	buf := a.rgbBuf
	for i, px := range a.Emu.Display.Pixels {
		v := byte(0)
		if px != 0 {
			v = 255
		}

		buf[i*3+0] = v
		buf[i*3+1] = v
		buf[i*3+2] = v
	}

	a.Texture.Update(nil, unsafe.Pointer(&buf[0]), a.Emu.Display.Width*3)
	a.Renderer.Clear()
	a.Renderer.Copy(a.Texture, nil, nil)
	a.Renderer.Present()
}

func handleKey(key sdl.Keycode, keypad *chip8.Keypad, down bool) {
	mapping := map[sdl.Keycode]byte{
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

	if k, ok := mapping[key]; ok {
		keypad.Keys[k] = down
	}
}
