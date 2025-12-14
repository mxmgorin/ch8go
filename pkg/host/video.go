package host

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

var DefaultPalette = Palette{
	Pixels: [16]Color{
		{0, 0, 0, 255},       // #000000
		{255, 255, 255, 255}, // #FFFFFF
		{170, 170, 170, 255}, // #AAAAAA
		{85, 85, 85, 255},    // #555555
		{255, 0, 0, 255},     // #FF0000
		{0, 255, 0, 255},     // #00FF00
		{0, 0, 255, 255},     // #0000FF
		{255, 255, 0, 255},   // #FFFF00
		{136, 0, 0, 255},     // #880000
		{0, 136, 0, 255},     // #008800
		{0, 0, 136, 255},     // #000088
		{136, 136, 0, 255},   // #888800
		{255, 0, 255, 255},   // #FF00FF
		{0, 255, 255, 255},   // #00FFFF
		{136, 0, 136, 255},   // #880088
		{0, 136, 136, 255},   // #008888
	},
	Buzzer:  Color{255, 255, 255, 255},
	Silence: Color{0, 0, 0, 255},
}

type Color [4]byte

func (c Color) ToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
}

type Palette struct {
	Pixels  [16]Color
	Buzzer  Color
	Silence Color
}

func (p *Palette) SetColor(index int, hex string) error {
	color, err := ParseHexColor(hex)

	if err != nil {
		return err
	}

	p.Pixels[index] = color

	return nil
}

func NewPalette(colors []string, buzzer, silence string) (p Palette, e error) {
	for i := 0; i < len(p.Pixels) && i < len(colors); i++ {
		p.SetColor(i, colors[i])
	}

	if buzzer != "" {
		color, err := ParseHexColor(buzzer)

		if err != nil {
			return p, err
		}

		p.Buzzer = color
	}

	if silence != "" {
		color, err := ParseHexColor(silence)

		if err != nil {
			return p, err
		}

		p.Buzzer = color
	}

	return p, nil
}

type FrameBuffer struct {
	Pixels     []byte
	SoundColor Color
	Width      int
	Height     int
	BPP        int
}

func newFrameBuffer(w, h, bpp int) FrameBuffer {
	return FrameBuffer{Pixels: make([]byte, w*h*bpp), Width: w, Height: h, BPP: bpp}
}

func (fb *FrameBuffer) Pitch() int {
	return fb.Width * fb.BPP
}

func (fb *FrameBuffer) Hash() string {
	sum := sha256.Sum256(fb.Pixels)
	return fmt.Sprintf("%x", sum[:])
}

func (fb *FrameBuffer) PNG() ([]byte, error) {
	if fb.BPP != 4 {
		return nil, fmt.Errorf("expected BPP=4 (RGBA), got %d", fb.BPP)
	}

	img := image.NewRGBA(image.Rect(0, 0, fb.Width, fb.Height))
	copy(img.Pix, fb.Pixels)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (fb *FrameBuffer) SavePNG(path string) error {
	data, err := fb.PNG()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (fb *FrameBuffer) Update(state chip8.FrameState, pal *Palette, display *chip8.Display) {
	if state.Dirty {
		pixelsCount := display.Size().Area()
		planes := display.Planes
		bpp := fb.BPP
		fbp := fb.Pixels
		palette := pal.Pixels

		for i := range pixelsCount {
			colorIdx := int(planes[0][i]) | int(planes[1][i])<<1 | int(planes[2][i])<<2 | int(planes[3][i])<<3
			idx := i * bpp
			copy(fbp[idx:idx+4], palette[colorIdx][:])
		}
	}

	if state.Beep {
		fb.SoundColor = pal.Buzzer
	} else {
		fb.SoundColor = pal.Silence
	}
}

func ParseHexColor(s string) (Color, error) {
	s = strings.TrimPrefix(s, "#")

	if len(s) != 6 {
		return Color{}, fmt.Errorf("invalid hex color: %q", s)
	}

	ri, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return Color{}, err
	}

	gi, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return Color{}, err
	}

	bi, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return Color{}, err
	}

	return Color{byte(ri), byte(gi), byte(bi), byte(255)}, nil
}
