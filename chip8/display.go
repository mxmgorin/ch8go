package chip8

import (
	"math/bits"
)

const (
	Chip8Width  = 64
	Chip8Height = 32
	SChipWidth  = 128
	SChipHeight = 64
	lowresScale = SChipWidth / Chip8Width
)

type Display struct {
	Planes        [4][]byte
	Width         int
	Height        int
	dirty         bool
	hires         bool
	pendingVBlank bool
	planeMask     int
}

func NewDisplay() Display {
	d := Display{}
	d.opRes(false)
	d.planeMask = 1
	d.Width = SChipWidth
	d.Height = SChipHeight

	for i := range d.Planes {
		d.Planes[i] = make([]byte, d.Width*d.Height)
		clear(d.Planes[i])
	}

	return d
}

func (d *Display) Reset() {
	d.opClear()
	d.opRes(false)
	d.pendingVBlank = false
	d.planeMask = 1
}

func (d *Display) poll() bool {
	result := d.dirty
	d.dirty = false
	d.pendingVBlank = false

	return result
}

func (d *Display) opClear() {
	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		for i := range d.Planes[plane] {
			d.Planes[plane][i] = 0
		}

		clear(d.Planes[plane])
	}

	d.dirty = true
}

func (d *Display) opPlane(x uint16) {
	d.planeMask = int(x)
}

func (d *Display) opRes(hires bool) {
	d.hires = hires
	for i := range d.Planes {
		clear(d.Planes[i])
	}
}

// Draws an N×H sprite where each row has `bytesPerRow` bytes.
// For example:
//   - Classic/SCHIP 8×H: bytesPerRow = 1
//   - XO-CHIP/SCHIP-16: bytesPerRow = 2
//
// Sprite layout across planes is planar:
//
//	plane 0 rows, plane 1 rows, etc.
//
// Returns collision count.
func (d *Display) DrawSprite(
	x, y byte,
	sprite []byte,
	width, height, bytesPerRow int,
	wrap bool,
) (collisions int) {
	if d.planeMask == 0 {
		return 0
	}

	d.pendingVBlank = true
	wrap = wrap || d.spriteWrap(int(x), int(y), width, height)

	baseX := int(x)
	baseY := int(y)

	// Precomputed masks (16 bits)
	var bitMask = [...]uint16{
		1 << 15, 1 << 14, 1 << 13, 1 << 12,
		1 << 11, 1 << 10, 1 << 9, 1 << 8,
		1 << 7, 1 << 6, 1 << 5, 1 << 4,
		1 << 3, 1 << 2, 1 << 1, 1 << 0,
	}

	planeStride := height * bytesPerRow
	planeIdx := 0

	for pi := range d.Planes {
		if d.isPlaneDisabled(pi) {
			continue
		}

		rowBase := planeIdx * planeStride
		planeIdx++

		for row := range height {
			var rowBits uint16
			if bytesPerRow == 1 {
				rowBits = uint16(sprite[rowBase+row]) << 8
			} else {
				off := rowBase + row*2
				rowBits = uint16(sprite[off])<<8 | uint16(sprite[off+1])
			}

			py := baseY + row

			for col := range width {
				if rowBits&bitMask[col] == 0 {
					continue
				}

				if d.togglePixelScaled(pi, baseX+col, py, wrap) {
					collisions++
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

	w := d.Width
	h := d.Height
	shift := n * w        // number of pixels to move downward
	remain := (h - n) * w // number of pixels that stay visible

	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		pixels := d.Planes[plane]
		// Move everything down using a single memmove
		copy(pixels[shift:shift+remain], pixels[:remain])
		// Clear top n rows
		clear(pixels[:shift])
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

		pixels := d.Planes[plane]

		for y := range h {
			row := y * w
			start := row
			end := row + w

			// Shift right: copy(src, dst) with overlap (safe: Go uses memmove)
			copy(pixels[start+n:end], pixels[start:start+(w-n)])
			// Clear leftmost n pixels
			clear(pixels[start : start+n])
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

			// scroll left using one copy
			copy(pixels[row:row+w-n], pixels[row+n:row+w])
			// clear rightmost n pixels
			clear(pixels[row+w-n : row+w])
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

	scroll := n * w
	remain := (h - n) * w

	for plane := range d.Planes {
		if d.isPlaneDisabled(plane) {
			continue
		}

		pixels := d.Planes[plane]
		// Move everything in one go
		copy(pixels[:remain], pixels[scroll:scroll+remain])
		// Clear bottom N rows
		clear(pixels[remain:])
	}
}

func (d *Display) isPlaneDisabled(plane int) bool {
	return d.planeMask&(1<<plane) == 0
}

func (d *Display) planesLen() int {
	return bits.OnesCount(uint(d.planeMask))
}
