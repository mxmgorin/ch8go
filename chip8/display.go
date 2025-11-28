package chip8

const lowresScale = 2

type Display struct {
	Pixels        []byte
	Width         int
	Height        int
	dirty         bool
	hires         bool
	pendingVBlank bool
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

	d.dirty = true
}

func (d *Display) Reset() {
	d.Clear()
	d.setResolution(false)
	d.pendingVBlank = false
}

func (d *Display) poll() bool {
	result := d.dirty
	d.dirty = false
	d.pendingVBlank = false

	return result
}

func (d *Display) setResolution(hires bool) {
	d.Width = 128
	d.Height = 64
	d.hires = hires
	d.Pixels = make([]byte, d.Width*d.Height)
}

// Draws a 8xN sprite at (x,y).
// Returns count of collisions occurred.
func (d *Display) DrawSprite(x, y byte, sprite []byte, wrap bool) (collisions int) {
	d.pendingVBlank = true
	w := 8
	h := len(sprite)
	wrap = wrap || d.spriteWrap(int(x), int(y), w, h)

	for row, b := range sprite {
		for col := range 8 {
			if b&(0x80>>col) == 0 {
				continue
			}

			if d.togglePixelScaled(
				int(x)+col,
				int(y)+row,
				wrap,
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
func (d *Display) DrawSprite16(x, y byte, sprite []byte, wrap bool) (collisions int) {
	d.pendingVBlank = true
	w := 16
	h := 16
	wrap = wrap || d.spriteWrap(int(x), int(y), w, h)

	for row := range 16 {
		hi := sprite[row*2]   // high byte
		lo := sprite[row*2+1] // low byte
		rowBits := uint16(hi)<<8 | uint16(lo)

		for col := range 16 {
			bit := (rowBits >> (15 - col)) & 1
			if bit == 1 {
				if d.togglePixelScaled(
					int(x)+col,
					int(y)+row,
					wrap,
				) {
					collisions += 1
				}
			}
		}
	}

	return collisions
}

func (d *Display) togglePixelScaled(x, y int, wrap bool) (collision bool) {
	if d.hires {
		return d.togglePixel(x, y, wrap)
	}

	// scale coordinates
	x *= lowresScale
	y *= lowresScale
	// toggle a 2x2 block
	collision = d.togglePixel(x, y, wrap) || collision
	collision = d.togglePixel(x+1, y, wrap) || collision
	collision = d.togglePixel(x, y+1, wrap) || collision
	collision = d.togglePixel(x+1, y+1, wrap) || collision
	return collision
}

// XORs the pixel at (x, y).
// Returns true if this operation turned a pixel off (collision).
func (d *Display) togglePixel(x, y int, wrap bool) bool {
	if wrap {
		// wrap around screen
		x = x % d.Width
		y = y % d.Height
	} else {
		// Clip: skip pixel if outside visible area
		if x < 0 || x >= d.Width || y < 0 || y >= d.Height {
			return false // no collision, nothing toggled
		}
	}
	i := y*d.Width + x
	pixel := d.Pixels[i]
	d.Pixels[i] ^= 1
	d.dirty = true

	return pixel == 1
}

// A sprite fully offscreen should wrap around screen
func (d *Display) spriteWrap(x, y, w, h int) bool {
	if !d.hires {
		x = x * lowresScale
		y = y * lowresScale
	}

	return d.spriteFullyOffscreen(x, y, w, h)
}

func (d *Display) spriteFullyOffscreen(x, y, w, h int) bool {
	return x+w <= 0 || x >= d.Width || y+h <= 0 || y >= d.Height
}

// schip extension
func (d *Display) ScrollDown(in byte, scale bool) {
	d.dirty = true
	n := int(in)
	if scale && !d.hires {
		n *= lowresScale
	}
	h := d.Height
	w := d.Width

	// Scroll from bottom upwards
	for y := h - 1; y >= n; y-- {
		copy(d.Pixels[y*w:y*w+w], d.Pixels[(y-n)*w:(y-n)*w+w])
	}

	// Clear new top rows
	for y := range n {
		for x := range w {
			d.Pixels[y*w+x] = 0
		}
	}
}

func (d *Display) ScrollRight4(scale bool) {
	d.dirty = true
	w := d.Width
	h := d.Height
	n := 4
	if scale && !d.hires {
		n *= lowresScale
	}

	for y := range h {
		row := y * w

		// scroll right
		for x := w - 1; x >= n; x-- {
			d.Pixels[row+x] = d.Pixels[row+(x-n)]
		}

		for x := range n {
			d.Pixels[row+x] = 0
		}
	}
}

func (d *Display) ScrollLeft4(scale bool) {
	d.dirty = true
	w := d.Width
	h := d.Height
	n := 4
	if scale && !d.hires {
		n *= lowresScale
	}

	for y := range h {
		row := y * w

		// shift left
		for x := 0; x < w-n; x++ {
			d.Pixels[row+x] = d.Pixels[row+(x+n)]
		}

		// clear rightmost 4 pixels
		for x := w - n; x < w; x++ {
			d.Pixels[row+x] = 0
		}
	}
}
