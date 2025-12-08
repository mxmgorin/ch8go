package chip8

import "testing"

func BenchmarkScrollUp(b *testing.B) {
	display := NewDisplay()

	b.ResetTimer()

	for b.Loop() {
		display.ScrollUp(16)
	}
}

func BenchmarkScrollDown(b *testing.B) {
	display := NewDisplay()

	b.ResetTimer()

	for b.Loop() {
		display.ScrollDown(16, true)
	}
}

func BenchmarkScrollLeft(b *testing.B) {
	display := NewDisplay()

	b.ResetTimer()

	for b.Loop() {
		display.ScrollLeft4(true)
	}
}

func BenchmarkScrollRight(b *testing.B) {
	display := NewDisplay()

	b.ResetTimer()

	for b.Loop() {
		display.ScrollRight4(true)
	}
}

func BenchmarkDrawSprite(b *testing.B) {
	display := NewDisplay()
	sprite := make([]byte, 16)

	b.ResetTimer()

	for b.Loop() {
		display.DrawSprite(0, 16, sprite, 8, 8, 1, false)
	}
}
