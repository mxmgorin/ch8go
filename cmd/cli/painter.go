package main

import (
	"strings"

	"github.com/mxmgorin/ch8go/pkg/host"
)

type ASCIIPainter struct{}

func (p *ASCIIPainter) Paint(fb *host.FrameBuffer) {
	const (
		on  = "██"
		off = "░░"
	)

	h := fb.Height
	w := fb.Width
	out := strings.Builder{}
	out.Grow(h * w * 2)

	for y := range h {
		for x := range w {
			i := (y*w + x) * fb.BPP
			r := fb.Pixels[i]
			g := fb.Pixels[i+1]
			b := fb.Pixels[i+2]

			if (r | g | b) != 0 {
				out.WriteString(on)
			} else {
				out.WriteString(off)
			}
		}

		out.WriteByte('\n')
	}

	ascii := out.String()
	println(ascii)
}
