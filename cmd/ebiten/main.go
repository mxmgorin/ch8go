package main

import (
	"log"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mxmgorin/ch8go/app"
)

var keymap = map[ebiten.Key]byte{
	ebiten.Key1: 0x1,
	ebiten.Key2: 0x2,
	ebiten.Key3: 0x3,
	ebiten.Key4: 0xC,

	ebiten.KeyQ: 0x4,
	ebiten.KeyW: 0x5,
	ebiten.KeyE: 0x6,
	ebiten.KeyR: 0xD,

	ebiten.KeyA: 0x7,
	ebiten.KeyS: 0x8,
	ebiten.KeyD: 0x9,
	ebiten.KeyF: 0xE,

	ebiten.KeyZ: 0xA,
	ebiten.KeyX: 0x0,
	ebiten.KeyC: 0xB,
	ebiten.KeyV: 0xF,
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
