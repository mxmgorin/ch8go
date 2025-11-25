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

// Draws a 8xN sprite at (x,y).
// Returns count of collisions occurred.
func (d *Display) DrawSprite(x, y byte, sprite []byte) (collisions int) {
	for row := range sprite {
		b := sprite[row]

		for col := range 8 {
			if b&(0x80>>col) == 0 {
				continue
			}

			if d.TogglePixel(
				int(x)+col,
				int(y)+row,
			) {
				collisions += 1
			}
		}
	}

	return collisions
}

// Draws a 16x16 Super-CHIP sprite.
// sprite is 32 bytes: 2 bytes per row × 16 rows.
// Returns count of collisions occurred.
func (d *Display) DrawSprite16(x, y byte, sprite []byte) (collisions int) {
	for row := range 16 {
		hi := sprite[row*2]   // high byte
		lo := sprite[row*2+1] // low byte
		rowBits := uint16(hi)<<8 | uint16(lo)

		for col := range 16 {
			bit := (rowBits >> (15 - col)) & 1
			if bit == 1 {
				if d.TogglePixel(
					int(x)+col,
					int(y)+row,
				) {
					collisions += 1
				}
			}
		}
	}

	return collisions
}

// XORs the pixel at (x, y).
// Returns true if this operation turned a pixel off (collision).
func (d *Display) TogglePixel(x, y int) bool {
	// wrap around screen
	x = x % d.Width
	y = y % d.Height
	i := y*d.Width + x
	pixel := d.Pixels[i]
	d.Pixels[i] ^= 1

	return pixel == 1
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
