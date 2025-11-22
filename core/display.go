package core

const (
	Width  = 64
	Height = 32
)

type Display struct {
	Pixels [Width * Height]byte
}

func NewDisplay() *Display {
	return &Display{}
}

func (d *Display) Clear() {
	for i := range d.Pixels {
		d.Pixels[i] = 0
	}
}
