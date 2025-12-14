package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mxmgorin/ch8go/pkg/chip8"
	"github.com/mxmgorin/ch8go/pkg/host"
)

var keymap = map[ebiten.Key]chip8.Key{
	ebiten.Key1: chip8.Key1,
	ebiten.Key2: chip8.Key2,
	ebiten.Key3: chip8.Key3,
	ebiten.Key4: chip8.KeyC,

	ebiten.KeyQ: chip8.Key4,
	ebiten.KeyW: chip8.Key5,
	ebiten.KeyE: chip8.Key6,
	ebiten.KeyR: chip8.KeyD,

	ebiten.KeyA: chip8.Key7,
	ebiten.KeyS: chip8.Key8,
	ebiten.KeyD: chip8.Key9,
	ebiten.KeyF: chip8.KeyE,

	ebiten.KeyZ: chip8.KeyA,
	ebiten.KeyX: chip8.Key0,
	ebiten.KeyC: chip8.KeyB,
	ebiten.KeyV: chip8.KeyF,
}

type App struct {
	*host.Emu
	scale int
}

func NewApp(scale int) (*App, error) {
	base, err := host.NewEmu()
	if err != nil {
		return nil, err
	}
	size := base.VM.Display.Size()

	ebiten.SetWindowSize(size.Width*scale, size.Height*scale)
	ebiten.SetWindowTitle("ch8go ebiten")

	return &App{
		Emu:   base,
		scale: scale,
	}, nil
}

func (a *App) Draw(screen *ebiten.Image) {
	screen.WritePixels(a.FrameBuffer.Pixels)
}

func (a *App) handleInput() {
	for k, v := range keymap {
		if ebiten.IsKeyPressed(k) {
			a.VM.Keypad.Press(v)
		} else {
			a.VM.Keypad.Release(v)
		}
	}
}

func (a *App) Update() error {
	a.handleInput()
	a.RunFrame()
	return nil
}

func (a *App) Layout(outsideW, outsideH int) (int, int) {
	size := a.VM.Display.Size()
	return size.Width, size.Height
}

func main() {
	slog.Info("ch8go ebiten")
	fs := flag.NewFlagSet("ch8go", flag.ExitOnError)
	opts, err := host.ParseOptions(fs, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := opts.ValidateROMPath(); err != nil {
		log.Fatal(err)
	}

	app, err := NewApp(opts.Scale)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := app.ReadROM(opts.ROMPath); err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
