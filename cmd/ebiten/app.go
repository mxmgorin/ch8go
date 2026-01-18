package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mxmgorin/ch8go/pkg/host"
)

type App struct {
	*host.Emu
	scale int
}

func newApp(scale int) (*App, error) {
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

func (a *App) Update() error {
	handleKeys(a)
	a.RunFrame()
	return nil
}

func (a *App) Layout(outsideW, outsideH int) (int, int) {
	size := a.VM.Display.Size()
	return size.Width, size.Height
}

func (a *App) run() error {
	return ebiten.RunGame(a)
}
