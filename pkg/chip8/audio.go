package chip8

import (
	"math"
)

const BeepFreq = 440.0 // or 400.0

// AudioMode selects the active sound generation model.
type AudioMode int

const (
	// AudioChip8 uses the classic CHIP-8 square wave sound.
	AudioChip8 AudioMode = iota
	// AudioXOChip uses the XO-CHIP programmable 16-byte audio pattern.
	AudioXOChip
)

// Audio represents the CHIP-8 / XO-CHIP sound generator state.
//
// It models the programmable audio pattern, pitch, and playback
// position used by the virtual machine. The Audio type contains
// no host-specific audio output logic and is controlled by the sound timer.
type Audio struct {
	// pattern is a 128-bit (16-byte) audio pattern buffer.
	// Each bit represents one step of the waveform.
	pattern [16]byte

	// pitch controls the playback frequency of the pattern.
	// In XO-CHIP, the default value is 64, corresponding to 4000 Hz.
	pitch byte

	// st is the sound timer. While ST > 0, audio playback is active.
	st byte

	// phase represents the current playback position within the pattern.
	// It is expressed as a fractional index to allow sub-step advancement.
	phase float64

	// mode selects the active audio mode (e.g. CHIP-8 or XO-CHIP).
	mode AudioMode
}

func NewAudio() Audio {
	a := Audio{}
	a.Reset()
	return a
}

func (a *Audio) Reset() {
	a.pattern = [16]byte{}
	a.pitch = 0
	a.phase = 0
	a.st = 0
	a.SetMode(AudioChip8)
}

func (a *Audio) TickTimer() bool {
	if a.st > 0 {
		a.st--
		return true
	}

	return false
}

func (a *Audio) SetMode(mode AudioMode) {
	a.mode = mode
	switch mode {
	case AudioXOChip:
		if a.pitch == 0 { // set default pitch for xochip
			a.pitch = 64
		}
	case AudioChip8:
		a.pitch = 0
	}
}

func (a *Audio) opPitch(xv byte) {
	a.pitch = xv
	a.SetMode(AudioXOChip)
}

func (a *Audio) opPattern(mem *Memory, addr uint16) {
	a.SetMode(AudioXOChip)

	for i := range a.pattern {
		a.pattern[i] = mem.Read(addr + uint16(i))
	}
}

func (a *Audio) Beep() bool {
	return a.st > 0
}

func (a *Audio) Output(out []float32, sampleRate float64) {
	if !a.Beep() {
		outputSilence(out)
		return
	}

	switch a.mode {
	case AudioXOChip:
		a.outputPattern(out, sampleRate)
	case AudioChip8:
		a.outputBeep(out, sampleRate)
	}
}

func (a *Audio) samplePattern(pos int) float32 {
	pos &= 127 // wrap around 0..127

	byteIndex := pos >> 3 // div on 8
	bitIndex := 7 - (pos & 7)

	if (a.pattern[byteIndex]>>bitIndex)&1 == 1 {
		return 1.0
	}
	return -1.0
}

// XO-HIP
func (a *Audio) outputPattern(out []float32, sampleRate float64) {
	for i := range out {
		pos := int(a.phase) & 128
		out[i] = a.samplePattern(pos)
		stepSize := patternFreq(float64(a.pitch)) / sampleRate
		a.phase += stepSize
		if a.phase >= 128 {
			a.phase -= 128
		}
	}
}

// Chip8
func (a *Audio) outputBeep(out []float32, sampleRate float64) {
	for i := range out {
		if a.phase < 1.0 {
			out[i] = 1
		} else {
			out[i] = -1
		}

		stepSize := BeepFreq / sampleRate
		a.phase += stepSize
		if a.phase >= 2 { // square wave period = 2
			a.phase -= 2
		}
	}
}

func outputSilence(out []float32) {
	for i := range out {
		out[i] = 0
	}
}

// XOCHIP pattern formula: 4000*2^((vx-64)/48)
func patternFreq(pitch float64) float64 {
	return 4000.0 * math.Pow(2.0, (pitch-64.0)/48.0)
}
