package main

import (
	"unsafe"

	"github.com/mxmgorin/ch8go/pkg/host"
	"github.com/veandco/go-sdl2/sdl"
)

type Painter struct {
	window   *sdl.Window
	texture  *sdl.Texture
	renderer *sdl.Renderer
	scale    int
}

func newPainter(width, height, scale int) (*Painter, error) {
	window, err := sdl.CreateWindow("ch8go SDL2",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		int32(width*scale),
		int32(height*scale),
		sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	p := Painter{}
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

func (p *Painter) Paint(fb *host.FrameBuffer) {
	p.texture.Update(nil, unsafe.Pointer(&fb.Pixels[0]), fb.Pitch())
	p.renderer.Clear()
	p.renderer.Copy(p.texture, nil, nil)
	p.renderer.Present()
}

func (p *Painter) Destroy() {
	p.texture.Destroy()
	p.renderer.Destroy()
	p.window.Destroy()
}
