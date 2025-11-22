package core

type Chip8 struct {
}

func NewChip8() *Chip8 {
	c := &Chip8{}

	return c
}

func (c *Chip8) LoadROM(bytes []byte) error {
	return nil
}

func (c *Chip8) Step() {
}
