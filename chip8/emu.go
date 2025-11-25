package chip8

import "fmt"

type EmuStatus int

const (
	StatusNoRom EmuStatus = iota
	StatusLoaded
)

type Emu struct {
	Status  EmuStatus
	Cpu     Cpu
	Memory  Memory
	Display Display
	Keypad  Keypad
	RomSize int
	cpuHz   float64
	fps     float64

	timerAccum float64
	cycleAccum float64
}

func NewEmu() *Emu {
	return &Emu{
		Cpu:     NewCpu(),
		Memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
		cpuHz:   500.0,
		fps:     60.0,
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
	e.Status = StatusLoaded
	e.RomSize = len(bytes)

	return nil
}

func (e *Emu) Step() {
	opcode := e.Cpu.fetch(&e.Memory)
	e.Cpu.execute(opcode, &e.Memory, &e.Display, &e.Keypad)
	e.tickTimers()
}

func (e *Emu) RunFrame() bool {
	e.cycleAccum += e.cpuHz / e.fps

	for e.cycleAccum >= 1.0 {
		e.cycleAccum -= 1.0
		e.Step()
	}

	return e.Display.pollDirty()
}

func (e *Emu) tickTimers() {
	e.timerAccum += 60.0 / e.cpuHz

	for e.timerAccum >= 1.0 {
		e.timerAccum -= 1.0
		e.Cpu.tickTimers()
	}
}

func (e *Emu) PeekNext() DisasmInfo {
	pc := e.Cpu.pc
	op := e.Memory.ReadU16(pc)
	asm := Disasm(op)
	return DisasmInfo{PC: pc, Op: op, Asm: asm}
}

func (e *Emu) Peek(n int) []DisasmInfo {
	emuCopy := *e
	results := make([]DisasmInfo, 0, n)

	for range n {
		pc := emuCopy.Cpu.pc

		if int(pc)+1 >= len(emuCopy.Memory.bytes) {
			break
		}

		op := emuCopy.Memory.ReadU16(pc)
		asm := Disasm(op)

		results = append(results, DisasmInfo{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})

		emuCopy.Step()
	}

	return results
}

func (e *Emu) DisasmRom() []DisasmInfo {
	start := ProgramStart
	end := ProgramStart + e.RomSize
	results := make([]DisasmInfo, 0, e.RomSize/OP_SIZE)

	for pc := uint16(start); pc < uint16(end); pc += OP_SIZE {
		op := e.Memory.ReadU16(pc)
		asm := Disasm(op)

		results = append(results, DisasmInfo{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})
	}

	return results
}
