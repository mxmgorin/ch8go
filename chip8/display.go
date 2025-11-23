package chip8

const (
	DisplayWidth  = 64
	DisplayHeight = 32
)

type Display struct {
	Pixels [DisplayWidth * DisplayHeight]byte
}

func NewDisplay() Display {
	return Display{}
}

func (d *Display) Clear() {
	for i := range d.Pixels {
		d.Pixels[i] = 0
	}
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

			px := (X + uint16(col)) % uint16(DisplayWidth)
			py := (Y + uint16(row)) % uint16(DisplayHeight)
			idx := py*uint16(DisplayWidth) + px

			if d.Pixels[idx] == 1 {
				collision = true
			}

			d.Pixels[idx] ^= 1
		}
	}

	return collision
}
