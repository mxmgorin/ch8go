package chip8

const lowresScale = 2

type Display struct {
	Planes        [2][]byte
	Width         int
	Height        int
	dirty         bool
	hires         bool
	pendingVBlank bool
	planeMask     int
}

func NewDisplay() Display {
	d := Display{}
	d.setResolution(false)
	d.planeMask = 1
	return d
}

func (d *Display) Clear() {
	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		for i := range d.Planes[plane] {
			d.Planes[plane][i] = 0
		}
	}

	d.dirty = true
}

func (d *Display) Reset() {
	d.Clear()
	d.setResolution(false)
	d.pendingVBlank = false
	d.planeMask = 1
}

func (d *Display) poll() bool {
	result := d.dirty
	d.dirty = false
	d.pendingVBlank = false

	return result
}

func (d *Display) selectPlanes(vx byte) {
	d.planeMask = int(vx & 0x3) // keep only lowest 2 bits
}

func (d *Display) setResolution(hires bool) {
	d.Width = 128
	d.Height = 64
	d.hires = hires
	d.Planes[0] = make([]byte, d.Width*d.Height)
	d.Planes[1] = make([]byte, d.Width*d.Height)
}

// Draws a 8xN sprite at (x,y).
// Returns count of collisions occurred.
func (d *Display) DrawSprite(x, y byte, sprite []byte, wrap bool) (collisions int) {
	if d.planeMask == 0 {
		return
	}

	d.pendingVBlank = true
	w := 8
	h := len(sprite) / d.planeMask
	wrap = wrap || d.spriteWrap(int(x), int(y), w, h)

	for pi := range d.Planes {
		if d.isPlaneDisabled(pi) {
			continue
		}

		planeOffset := pi * h

		for row := range h {
			b := sprite[planeOffset+row]

			for col := range 8 {
				if b&(0x80>>col) == 0 {
					continue
				}

				if d.togglePixelScaled(
					pi,
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

// Draws a 16x16 Super-CHIP sprite.
// sprite is 32 bytes: 2 bytes per row × 16 rows.
// Returns count of collisions occurred.
func (d *Display) DrawSprite16(x, y byte, sprite []byte, wrap bool) (collisions int) {
	d.pendingVBlank = true
	w := 16
	h := 16
	wrap = wrap || d.spriteWrap(int(x), int(y), w, h)
	const bytesPerPlane = 32 // 16 rows × 2 bytes

	for pi := range d.Planes {
		if d.isPlaneDisabled(pi) {
			continue
		}

		planeOffset := pi * bytesPerPlane

		for row := range h {
			hi := sprite[planeOffset+row*2]
			lo := sprite[planeOffset+row*2+1]
			rowBits := uint16(hi)<<8 | uint16(lo)

			for col := range w {
				bit := (rowBits >> (15 - col)) & 1
				if bit != 1 {
					continue
				}

				if d.togglePixelScaled(
					pi,
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

func (d *Display) togglePixelScaled(plane, x, y int, wrap bool) (collision bool) {
	if d.hires {
		return d.togglePixel(plane, x, y, wrap)
	}

	// scale coordinates
	x *= lowresScale
	y *= lowresScale
	// toggle a 2x2 block
	collision = d.togglePixel(plane, x, y, wrap) || collision
	collision = d.togglePixel(plane, x+1, y, wrap) || collision
	collision = d.togglePixel(plane, x, y+1, wrap) || collision
	collision = d.togglePixel(plane, x+1, y+1, wrap) || collision
	return collision
}

// XORs the pixel at (x, y).
// Returns true if this operation turned a pixel off (collision).
func (d *Display) togglePixel(plane, x, y int, wrap bool) bool {
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
	pixel := d.Planes[plane][i]
	d.Planes[plane][i] ^= 1
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

	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		// Scroll from bottom upwards
		for y := h - 1; y >= n; y-- {
			copy(d.Planes[plane][y*w:y*w+w], d.Planes[plane][(y-n)*w:(y-n)*w+w])
		}

		// Clear new top rows
		for y := range n {
			for x := range w {
				d.Planes[plane][y*w+x] = 0
			}
		}
	}
}

// schip
func (d *Display) ScrollRight4(scale bool) {
	d.dirty = true
	w := d.Width
	h := d.Height
	n := 4
	if scale && !d.hires {
		n *= lowresScale
	}

	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		for y := range h {

			row := y * w

			// scroll right
			for x := w - 1; x >= n; x-- {
				d.Planes[plane][row+x] = d.Planes[plane][row+(x-n)]
			}

			for x := range n {
				d.Planes[plane][row+x] = 0
			}
		}
	}
}

// schip
func (d *Display) ScrollLeft4(scale bool) {
	d.dirty = true
	w := d.Width
	h := d.Height

	n := 4
	if scale && !d.hires {
		n *= lowresScale
	}

	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		pixels := d.Planes[plane]

		for y := range h {
			row := y * w

			// shift left
			for x := 0; x < w-n; x++ {
				pixels[row+x] = pixels[row+(x+n)]
			}

			// clear rightmost n pixels
			for x := w - n; x < w; x++ {
				pixels[row+x] = 0
			}
		}
	}
}

// xochip
func (d *Display) ScrollUp(n int) {
	if n <= 0 {
		return
	}

	if !d.hires {
		n *= lowresScale
	}

	w := d.Width
	h := d.Height
	d.dirty = true

	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		// Move rows upward
		for y := 0; y < h-n; y++ {
			src := (y + n) * w
			dst := y * w
			copy(d.Planes[plane][dst:dst+w], d.Planes[plane][src:src+w])
		}

		// Clear bottom n rows
		start := (h - n) * w
		for i := start; i < len(d.Planes[plane]); i++ {
			d.Planes[plane][i] = 0
		}
	}
}

func (d *Display) isPlaneDisabled(plane int) bool {
	return d.planeMask&(1<<plane) == 0
}
