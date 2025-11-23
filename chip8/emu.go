package chip8

import "fmt"

type Emu struct {
	Cpu     *Cpu
	memory  *Memory
	Display *Display
	Keypad  *Keypad
}

func NewEmu() *Emu {
	return &Emu{
		Cpu:     NewCpu(),
		memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
	}
}

func (cpu *Emu) LoadRom(bytes []byte) error {
	err := cpu.memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	return nil
}

func (cpu *Emu) Step() {
	cpu.Cpu.Step(cpu.memory, cpu.Display, cpu.Keypad)
}
