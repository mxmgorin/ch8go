package chip8

import "fmt"

type Emu struct {
	cpu     *Cpu
	memory  *Memory
	Display *Display
	Keypad  *Keypad
}

func NewEmu() Emu {
	return Emu{
		cpu:     NewCpu(),
		memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
	}
}

func (c *Emu) LoadRom(bytes []byte) error {
	err := c.memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	return nil
}

func (c *Emu) Step() {
	c.cpu.Step(c.memory, c.Display, c.Keypad)
}
