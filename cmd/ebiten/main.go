package main

import (
	"log"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mxmgorin/ch8go/app"
	"github.com/mxmgorin/ch8go/chip8"
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

type EbitenApp struct {
	*app.App
	scale int
}

func NewApp(scale int) (*EbitenApp, error) {
	base, err := app.NewApp()
	if err != nil {
		return nil, err
	}
	w := base.VM.Display.Width
	h := base.VM.Display.Height

	ebiten.SetWindowSize(w*scale, h*scale)
	ebiten.SetWindowTitle("ch8go ebiten")

	return &EbitenApp{
		App:   base,
		scale: scale,
	}, nil
}

func (a *EbitenApp) Draw(screen *ebiten.Image) {
	screen.WritePixels(a.FrameBuffer.Pixels)
}

func (a *EbitenApp) handleInput() {
	for k, v := range keymap {
		if ebiten.IsKeyPressed(k) {
			a.VM.Keypad.Press(v)
		} else {
			a.VM.Keypad.Release(v)
		}
	}
}

func (a *EbitenApp) Update() error {
	a.handleInput()
	a.RunFrame()
	return nil
}

func (a *EbitenApp) Layout(outsideW, outsideH int) (int, int) {
	return a.VM.Display.Width, a.VM.Display.Height
}

func main() {
	slog.Info("ch8go ebiten")
	romPath, scale := app.ParseFlags()

	a, err := NewApp(scale)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := a.ReadROM(romPath); err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}
