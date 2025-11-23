package core

import "fmt"

type Chip8 struct {
	cpu     *Cpu
	memory  *Memory
	Display *Display
	Keypad  *Keypad
}

func NewChip8() Chip8 {
	return Chip8{
		cpu:     NewCpu(),
		memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
	}
}

func (c *Chip8) LoadRom(bytes []byte) error {
	err := c.memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	return nil
}

func (c *Chip8) Step() {
	c.cpu.Step(c.memory, c.Display, c.Keypad)
}
