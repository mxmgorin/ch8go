package chip8

import (
	"math"
	"testing"
)

func TestAudioTickTimerAndBeep(t *testing.T) {
	a := NewAudio()
	if a.Beep() || a.TickTimer() {
		t.Error("fresh audio should not beep and TickTimer should be false")
	}

	a.st = 2
	if !a.Beep() {
		t.Error("Beep should be true while st > 0")
	}
	if !a.TickTimer() || a.st != 1 {
		t.Errorf("TickTimer should return true and decrement st, got st=%d", a.st)
	}
}

func TestAudioSetModeAndOps(t *testing.T) {
	a := NewAudio() // starts in CHIP-8 mode, pitch 0

	a.SetMode(AudioXOChip)
	if a.pitch != 64 {
		t.Errorf("XO-CHIP default pitch = %d, want 64", a.pitch)
	}

	a.pitch = 30
	a.SetMode(AudioXOChip) // a non-zero pitch must be preserved
	if a.pitch != 30 {
		t.Errorf("SetMode should preserve non-zero pitch, got %d", a.pitch)
	}

	a.SetMode(AudioChip8)
	if a.pitch != 0 {
		t.Errorf("CHIP-8 mode should reset pitch to 0, got %d", a.pitch)
	}

	a.opPitch(100)
	if a.pitch != 100 || a.mode != AudioXOChip {
		t.Errorf("opPitch: pitch=%d mode=%d, want 100/XO-CHIP", a.pitch, a.mode)
	}
}

func TestAudioOpPattern(t *testing.T) {
	vm := NewVM()
	for i := 0; i < 16; i++ {
		vm.Memory.bytes[0x300+i] = byte(i + 1)
	}

	vm.Audio.opPattern(&vm.Memory, 0x300)

	if vm.Audio.mode != AudioXOChip {
		t.Error("opPattern should switch to XO-CHIP mode")
	}
	for i := 0; i < 16; i++ {
		if vm.Audio.pattern[i] != byte(i+1) {
			t.Fatalf("pattern[%d] = %d, want %d", i, vm.Audio.pattern[i], i+1)
		}
	}
}

func TestSamplePattern(t *testing.T) {
	a := NewAudio()
	a.pattern[0] = 0x80 // only the top bit is set

	if a.samplePattern(0) != 1.0 {
		t.Error("a set bit should sample as 1.0")
	}
	if a.samplePattern(1) != -1.0 {
		t.Error("an unset bit should sample as -1.0")
	}
	if a.samplePattern(128) != a.samplePattern(0) {
		t.Error("position should wrap modulo 128")
	}
}

func TestPatternFreq(t *testing.T) {
	if got := patternFreq(64); math.Abs(got-4000.0) > 1e-6 {
		t.Errorf("patternFreq(64) = %v, want 4000", got)
	}
	if got := patternFreq(112); math.Abs(got-8000.0) > 1e-6 { // +48 => one octave up
		t.Errorf("patternFreq(112) = %v, want 8000", got)
	}
}

func TestAudioOutput(t *testing.T) {
	a := NewAudio()

	// Not beeping -> silence overwrites the buffer with zeros.
	out := make([]float32, 8)
	for i := range out {
		out[i] = 0.5
	}
	a.Output(out, 44100)
	for i, s := range out {
		if s != 0 {
			t.Fatalf("silence: out[%d] = %v, want 0", i, s)
		}
	}

	// CHIP-8 square wave over a full period produces both +1 and -1.
	a.st = 1
	a.SetMode(AudioChip8)
	a.phase = 0
	beep := make([]float32, 300)
	a.Output(beep, 44100)
	var hasHigh, hasLow bool
	for _, s := range beep {
		switch s {
		case 1:
			hasHigh = true
		case -1:
			hasLow = true
		}
	}
	if !hasHigh || !hasLow {
		t.Error("CHIP-8 beep should produce both +1 and -1 samples")
	}

	// XO-CHIP: an all-ones pattern yields +1 for every sample; the long
	// buffer also exercises the phase wrap-around.
	a.st = 1
	a.SetMode(AudioXOChip)
	for i := range a.pattern {
		a.pattern[i] = 0xFF
	}
	a.phase = 0
	pat := make([]float32, 1500)
	a.Output(pat, 44100)
	for i, s := range pat {
		if s != 1.0 {
			t.Fatalf("pattern output: pat[%d] = %v, want 1.0", i, s)
		}
	}
}

func TestAudioOutputPatternScans(t *testing.T) {
	a := NewAudio()
	a.st = 1
	a.SetMode(AudioXOChip) // pitch 64

	// Steps 0..7 sample as +1, steps 8..15 as -1. As playback advances the
	// output must move from +1 into -1 — a stuck position would stay all +1.
	a.pattern[0] = 0xFF
	a.pattern[1] = 0x00
	a.phase = 0

	out := make([]float32, 256)
	a.Output(out, 44100)

	var hasHigh, hasLow bool
	for _, s := range out {
		switch s {
		case 1:
			hasHigh = true
		case -1:
			hasLow = true
		}
	}
	if !hasHigh || !hasLow {
		t.Error("pattern playback position should scan across steps (got stuck)")
	}
}
