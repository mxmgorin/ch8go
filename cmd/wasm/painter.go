package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"syscall/js"

	"github.com/mxmgorin/ch8go/pkg/host"
)

type Painter struct {
	ctx           js.Value
	imageData     js.Value
	screen        js.Value
	canvas        js.Value
	screenBgColor *host.Color
	width         int
	height        int
}

func (p *Painter) setScreenBg(color host.Color) {
	if p.screenBgColor == nil || *p.screenBgColor != color {
		hex := color.ToHex()
		p.screen.Get("style").Set("background", hex)
		p.screenBgColor = &color
	}
}

func newPainter(w, h int) (Painter, error) {
	p := Painter{}
	p.width = w
	p.height = h
	doc := js.Global().Get("document")

	p.canvas = doc.Call("getElementById", "chip8-canvas")
	p.canvas.Set("width", w)
	p.canvas.Set("height", h)

	p.screen = doc.Call("getElementById", "chip8-screen")

	scaleInput := doc.Call("getElementById", "scaleInput")
	value := scaleInput.Get("value").String()
	p.setScale(value)
	scaleInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := scaleInput.Get("value").String()
		p.setScale(value)
		return nil
	}))

	p.ctx = p.canvas.Call("getContext", "2d")
	p.imageData = p.ctx.Call("createImageData", w, h)
	return p, nil
}

func (p *Painter) setScale(value string) {
	scale, err := strconv.Atoi(value)
	if err != nil {
		slog.Error("Invalid scale:", "scale", value)
		return
	}

	canvasStyle := p.canvas.Get("style")
	canvasStyle.Set("transform", fmt.Sprintf("scale(%d)", scale))

	screenStyle := p.screen.Get("style")
	screenStyle.Set("width", fmt.Sprintf("%dpx", p.width*scale))
	screenStyle.Set("height", fmt.Sprintf("%dpx", p.height*scale))
}

func (p *Painter) Paint(fb *host.FrameBuffer) {
	p.setScreenBg(fb.SoundColor)
	js.CopyBytesToJS(p.imageData.Get("data"), fb.Pixels)
	p.ctx.Call("putImageData", p.imageData, 0, 0)
}
