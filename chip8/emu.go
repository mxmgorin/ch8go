package chip8

import "fmt"

type Emu struct {
	Cpu     Cpu
	Memory  Memory
	Display Display
	Keypad  Keypad
}

func NewEmu() *Emu {
	return &Emu{
		Cpu:     NewCpu(),
		Memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
	}
}

func (e *Emu) LoadRom(bytes []byte) error {
	err := e.Memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}
	e.Display.Clear()
	e.Keypad.Reset()
	e.Cpu.Reset()

	return nil
}

func (e *Emu) Step() {
	opcode := e.Cpu.fetch(&e.Memory)
	e.Cpu.execute(opcode, &e.Memory, &e.Display, &e.Keypad)
	e.Cpu.tickTimers()
}

func (e *Emu) DisasmNext() DisasmInfo {
	pc := e.Cpu.pc
	op := e.Memory.ReadU16(pc)
	asm := DisasmOp(op)
	return DisasmInfo{PC: pc, Op: op, Asm: asm}
}

func (e *Emu) DisasmN(n int) []DisasmInfo {
	emuCopy := *e
	results := make([]DisasmInfo, 0, n)

	for range n {
		pc := emuCopy.Cpu.pc

		if int(pc)+1 >= len(emuCopy.Memory.bytes) {
			break
		}

		op := emuCopy.Memory.ReadU16(pc)
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
