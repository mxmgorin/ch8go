package app

import (
	"testing"
)

func BenchmarkRunFrame(b *testing.B) {
	dt := 1.0 / 60
	app, _ := NewApp()
	path := "../roms/chip8archive/danm8ku.ch8"
	if _, err := app.ReadROM(path); err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for b.Loop() {
		app.RunFrameDT(dt)
	}
}
