package chip8

import (
	"math"
)

const BeepFreq = 440.0 // or 400.0

type AudioMode int

const (
	AudioChip8  AudioMode = iota // normal square wave (CHIP-8 default)
	AudioXOChip                  // XO-Chip 16-byte pattern
)

type Audio struct {
	pattern [16]byte // 128 bits
	pitch   byte     // default in xochip is 64 meaning 4000 Hz
	st      byte
	phase   float64 // position inside pattern
	mode    AudioMode
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

func (a *Audio) OpPitch(xv byte) {
	a.pitch = xv
	a.SetMode(AudioXOChip)
}

func (a *Audio) OpPattern(mem *Memory, addr uint16) {
	a.SetMode(AudioXOChip)

	for i := range a.pattern {
		a.pattern[i] = mem.Read(addr + uint16(i))
	}
}

func (a *Audio) Beep() bool {
	return a.st > 0
}

func (a *Audio) Output(out []float32, freq float64) {
	if !a.Beep() {
		outputSilence(out)
		return
	}

	switch a.mode {
	case AudioXOChip:
		a.outputPattern(out, freq)
	case AudioChip8:
		a.outputBeep(out, freq)
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
func (a *Audio) outputPattern(out []float32, freq float64) {
	for i := range out {
		pos := int(a.phase) % 128
		out[i] = a.samplePattern(pos)
		stepSize := calcPlaybackRate(float64(a.pitch)) / freq
		a.phase += stepSize
		if a.phase >= 128 {
			a.phase -= 128
		}
	}
}

// Chip8
func (a *Audio) outputBeep(out []float32, freq float64) {
	for i := range out {
		if int(a.phase)&1 == 0 {
			out[i] = 1
		} else {
			out[i] = -1
		}

		stepSize := BeepFreq / freq
		a.phase += stepSize
		if a.phase >= 2 { // square wave period = 2
			a.phase -= 2
		}
	}
}

func (a *Audio) patternIsEmpty() bool {
	for _, b := range a.pattern {
		if b != 0 {
			return false
		}
	}
	return true
}

func outputSilence(out []float32) {
	for i := range out {
		out[i] = 0
	}
}

// XOCHIP pattern formula: 4000*2^((vx-64)/48)
func calcPlaybackRate(pitch float64) float64 {
	return 4000.0 * math.Pow(2.0, (pitch-64.0)/48.0)
}
