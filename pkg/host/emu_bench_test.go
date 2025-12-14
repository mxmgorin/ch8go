package host

import (
	"testing"
	"time"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

func BenchmarkRunFrame(b *testing.B) {
	elapsed := time.Second / 60
	app, _ := NewEmu()
	path := "../../testdata/roms/chip8archive/danm8ku.ch8"
	if _, err := app.ReadROM(path); err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for b.Loop() {
		app.runFrame(elapsed)
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
