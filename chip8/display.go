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
