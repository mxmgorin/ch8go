package chip8

type Display struct {
	Pixels []byte
	Width  int
	Height int
}

func NewDisplay() Display {
	d := Display{}
	d.setResolution(false)
	return d
}

func (d *Display) Clear() {
	for i := range d.Pixels {
		d.Pixels[i] = 0
	}
}

func (d *Display) setResolution(hires bool) {
	if hires {
		d.Width = 128
		d.Height = 64
	} else {
		d.Width = 64
		d.Height = 32
	}

	d.Pixels = make([]byte, d.Width*d.Height)
}

// DrawSprite draws N rows of sprite data at (x,y).
// Returns true if any pixel was unset (collision).
func (d *Display) DrawSprite(x, y byte, sprite []byte) (collision bool) {
	collision = false
	X := uint16(x)
	Y := uint16(y)

	for row := range sprite {
		b := sprite[row]

		for col := range 8 {
			if b&(0x80>>col) == 0 {
				continue
			}

			px := (X + uint16(col)) % uint16(d.Width)
			py := (Y + uint16(row)) % uint16(d.Height)
			idx := py*uint16(d.Width) + px

			if d.Pixels[idx] == 1 {
				collision = true
			}

			d.Pixels[idx] ^= 1
		}
	}

	return collision
}

// schip extension
func (d *Display) ScrollDown(n byte) {
	N := int(n)
	h := d.Height
	w := d.Width

	// Scroll from bottom upwards
	for y := h - 1; y >= N; y-- {
		copy(d.Pixels[y*w:y*w+w], d.Pixels[(y-N)*w:(y-N)*w+w])
	}

	// Clear new top rows
	for y := range N {
		for x := range w {
			d.Pixels[y*w+x] = 0
		}
	}
}

func (d *Display) ScrollRight4() {
	w := d.Width
	h := d.Height

	for y := range h {
		row := y * w

		// scroll right
		for x := w - 1; x >= 4; x-- {
			d.Pixels[row+x] = d.Pixels[row+(x-4)]
		}

		for x := range 4 {
			d.Pixels[row+x] = 0
		}
	}
}

func (d *Display) ScrollLeft4() {
	w := d.Width
	h := d.Height

	for y := range h {
		row := y * w

		// shift left
		for x := 0; x < w-4; x++ {
			d.Pixels[row+x] = d.Pixels[row+(x+4)]
		}

		// clear rightmost 4 pixels
		for x := w - 4; x < w; x++ {
			d.Pixels[row+x] = 0
		}
	}
}
