package chip8

import "math"

type Audio struct {
	pattern [16]byte // 128 bits
	pitch   byte     // default 64 meaning 4000 Hz
	phase   float64  // position inside pattern
	st      byte
}

func NewAudio() Audio {
	a := Audio{}
	a.Reset()
	return a
}

func (a *Audio) Reset() {
	a.pattern = [16]byte{}
	a.pitch = 64 // default playback if 4000
	a.st = 0
}

func (a *Audio) tickTimer() {
	if a.st > 0 {
		a.st--
	}
}

func (a *Audio) Beep() bool {
	return a.st > 0
}

// 4000*2^((vx-64)/48)
func (a *Audio) PlaybackRate() float64 {
	pitch := float64(a.pitch)
	return 4000.0 * math.Pow(2.0, (pitch-64.0)/48.0)
}

func (a *Audio) Sample(pos int) float32 {
	pos &= 127 // wrap around 0..127

	byteIndex := pos >> 3 // div on 8
	bitIndex := 7 - (pos & 7)

	if (a.pattern[byteIndex]>>bitIndex)&1 == 1 {
		return 1.0
	}
	return -1.0
}

func (a *Audio) Output(out []float32, freq float64) {
	if a.pitch == 0 || !a.Beep() {
		silence(out)
		return
	}

	step := a.PlaybackRate() / freq

	for i := range out {
		pos := int(a.phase) % 128
		out[i] = a.Sample(pos)
		a.phase += step
		if a.phase >= 128 {
			a.phase -= 128
		}
	}
}

func (a *Audio) outputChip8(out []float32, freq float64) {
	if !a.Beep() {
		silence(out)
		return
	}

	beepFreq := 440.0 // or 400.0, up to you
	step := beepFreq / freq

	for i := range out {
		if int(a.phase)&1 == 0 {
			out[i] = 1
		} else {
			out[i] = -1
		}

		a.phase += step
		if a.phase >= 2 { // square wave period = 2
			a.phase -= 2
		}
	}
}

func silence(out []float32) {
	for i := range out {
		out[i] = 0
	}
}
