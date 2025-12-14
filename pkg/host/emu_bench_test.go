package host

import (
	"testing"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

func BenchmarkRunFrame(b *testing.B) {
	dt := 1.0 / 60
	app, _ := NewEmu()
	path := "../../testdata/roms/chip8archive/danm8ku.ch8"
	if _, err := app.ReadROM(path); err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for b.Loop() {
		app.RunFrameDT(dt)
	}
}

func BenchmarkUpdateFrameBuffer(b *testing.B) {
	fs := chip8.FrameState{
		Dirty: true,
	}
	app, _ := NewEmu()
	fb := &app.FrameBuffer

	b.ResetTimer()

	for b.Loop() {
		fb.Update(fs, &app.Palette, &app.VM.Display)
	}
}
