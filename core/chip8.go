package core

import "fmt"

type Chip8 struct {
	cpu     *Cpu
	memory  *Memory
	display *Display
	keypad  *Keypad
}

func NewChip8() *Chip8 {
	return &Chip8{
		cpu:     NewCpu(),
		memory:  NewMemory(),
		display: NewDisplay(),
		keypad:  NewKeypad(),
	}
}

func (c *Chip8) LoadROM(bytes []byte) error {
	err := c.memory.LoadRom(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	return nil
}

func (c *Chip8) Step() {
	c.cpu.Step(c.memory, c.display, c.keypad)
}
