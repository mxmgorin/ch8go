package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"unsafe"
	"log"

	"github.com/mxmgorin/ch8go/core"
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

	chip8 := core.NewChip8()
	if err := chip8.LoadRom(rom); err != nil {
		panic(err)
	}

	fmt.Println("CHIP-8 SDL2")
	fmt.Printf("ROM: %s\n", *romPath)
	fmt.Printf("Speed: %d Hz\n", *hz)

	app := newApp()
	app.run(chip8, hz)
	app.quit()
}

type App struct {
	Window   *sdl.Window
	Texture  *sdl.Texture
	Renderer *sdl.Renderer
}

func newApp() App {
	app := App{}
	app.init()
	return app
}

func (app *App) init() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow(
		"CHIP-8",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		core.DisplayWidth*WindowScale,
		core.DisplayHeight*WindowScale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}
	app.Window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	app.Renderer = renderer

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24,
		sdl.TEXTUREACCESS_STREAMING,
		core.DisplayWidth,
		core.DisplayHeight,
	)
	if err != nil {
		panic(err)
	}

	app.Texture = texture
}

func (app *App) quit() {
	sdl.Quit()
	app.Window.Destroy()
	app.Renderer.Destroy()
	app.Texture.Destroy()
}

func (app *App) run(chip8 *core.Chip8, hz *int) {
	cycleDelay := time.Second / time.Duration(*hz)

	running := true
	for running {
		frameStart := time.Now()

		// Handle events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				switch ev.Type {
				case sdl.KEYDOWN:
					handleKey(ev.Keysym.Sym, chip8.Keypad, true)
				case sdl.KEYUP:
					handleKey(ev.Keysym.Sym, chip8.Keypad, false)
				}
			}
		}

		chip8.Step()
		app.draw(chip8.Display)

		elapsed := time.Since(frameStart)
		if elapsed < cycleDelay {
			time.Sleep(cycleDelay - elapsed)
		}
	}
}

func (app *App) draw(d *core.Display) {
	// Convert pixel array -> RGB buffer
	buf := make([]byte, len(d.Pixels)*3)

	for i, px := range d.Pixels {
		v := byte(0)
		if px != 0 {
			v = 255
		}

		buf[i*3+0] = v
		buf[i*3+1] = v
		buf[i*3+2] = v
	}

	app.Texture.Update(nil, unsafe.Pointer(&buf[0]), core.DisplayWidth*3)
	app.Renderer.Clear()
	app.Renderer.Copy(app.Texture, nil, nil)
	app.Renderer.Present()
}

func handleKey(key sdl.Keycode, keypad *core.Keypad, down bool) {
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
