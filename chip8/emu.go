package chip8

import "fmt"

type Emu struct {
	Cpu     Cpu
	memory  Memory
	Display Display
	Keypad  Keypad
}

func NewEmu() *Emu {
	return &Emu{
		Cpu:     NewCpu(),
		memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
	}
}

func (e *Emu) LoadRom(bytes []byte) error {
	err := e.memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}
	e.Display.Clear()
	e.Keypad.Reset()
	e.Cpu.Reset()

	return nil
}

func (e *Emu) Step() {
	e.Cpu.Step(&e.memory, &e.Display, &e.Keypad)
	e.Cpu.UpdateTimers()
}

func (e *Emu) DisasmNext() DisasmInfo {
	pc := e.Cpu.pc
	op := e.memory.ReadU16(pc)
	asm := DisasmOp(op)
	return DisasmInfo{PC: pc, Op: op, Asm: asm}
}

func (e *Emu) DisasmN(n int) []DisasmInfo {
	emuCopy := *e
	results := make([]DisasmInfo, 0, n)

	for range n {
		pc := emuCopy.Cpu.pc

		if int(pc)+1 >= len(emuCopy.memory.bytes) {
			break
		}

		op := emuCopy.memory.ReadU16(pc)
		asm := DisasmOp(op)

		results = append(results, DisasmInfo{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})

		emuCopy.Step()
	}

	return results
}
