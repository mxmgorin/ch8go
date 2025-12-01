package chip8

import (
	"math"
)

const BeepFreq = 440.0 // or 400.0
const OutputFreq = 44100.0

type Audio struct {
	pattern  [16]byte // 128 bits
	pitch    byte     // default in xochip is 64 meaning 4000 Hz
	st       byte
	phase    float64 // position inside pattern
	stepSize float64
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
	a.stepSize = BeepFreq / OutputFreq
}

func (a *Audio) TickTimer() {
	if a.st > 0 {
		a.st--
	}
}

func (a *Audio) SetPitch(xv byte) {
	a.pitch = xv
	a.stepSize = a.playbackRate()
}

func (a *Audio) LoadPattern(mem *Memory, addr uint16) {
	if a.pitch == 0 { // set default pitch for xochip
		a.pitch = 64
	}

	for i := range a.pattern {
		a.pattern[i] = mem.Read(addr + uint16(i))
	}
}

func (a *Audio) Beep() bool {
	return a.st > 0
}

func (a *Audio) Output(out []float32) {
	if !a.Beep() {
		outputSilence(out)
		return
	}

	if a.patternIsEmpty() { // assume it is chip8 when pattern not used
		a.outputBeep(out)
	} else {
		a.outputPattern(out)
	}
}

// 4000*2^((vx-64)/48)
func (a *Audio) playbackRate() float64 {
	pitch := float64(a.pitch)
	return 4000.0 * math.Pow(2.0, (pitch-64.0)/48.0)
}

func (a *Audio) sample(pos int) float32 {
	pos &= 127 // wrap around 0..127

	byteIndex := pos >> 3 // div on 8
	bitIndex := 7 - (pos & 7)

	if (a.pattern[byteIndex]>>bitIndex)&1 == 1 {
		return 1.0
	}
	return -1.0
}

// XO-HIP
func (a *Audio) outputPattern(out []float32) {
	for i := range out {
		pos := int(a.phase) % 128
		out[i] = a.sample(pos)
		a.phase += a.stepSize
		if a.phase >= 128 {
			a.phase -= 128
		}
	}
}

// Chip8
func (a *Audio) outputBeep(out []float32) {
	for i := range out {
		if int(a.phase)&1 == 0 {
			out[i] = 1
		} else {
			out[i] = -1
		}

		a.phase += a.stepSize
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
