package chip8

import "fmt"

type EmuStatus int

const (
	StatusNoRom EmuStatus = iota
	StatusLoaded
)

type VM struct {
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

func NewVM() *VM {
	return &VM{
		Cpu:     NewCpu(QuirksSuperChip11),
		Memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
		cpuHz:   600.0,
		fps:     60.0,
	}
}

func (vm *VM) LoadRom(bytes []byte) error {
	err := vm.Memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}
	vm.Display.Clear()
	vm.Keypad.Reset()
	vm.Cpu.Reset()
	vm.Status = StatusLoaded
	vm.RomSize = len(bytes)

	return nil
}

func (vm *VM) Step() {
	if !vm.Display.pendingVBlank || !vm.Cpu.quirks.VBlankWait {
		opcode := vm.Cpu.fetch(&vm.Memory)
		vm.Cpu.execute(opcode, &vm.Memory, &vm.Display, &vm.Keypad)
	}

	vm.tickTimers()
}

func (vm *VM) RunFrame() bool {
	vm.cycleAccum += vm.cpuHz / vm.fps

	for vm.cycleAccum >= 1.0 {
		vm.cycleAccum -= 1.0
		vm.Step()
	}

	return vm.Display.poll()
}

func (vm *VM) tickTimers() {
	vm.timerAccum += 60.0 / vm.cpuHz

	for vm.timerAccum >= 1.0 {
		vm.timerAccum -= 1.0
		vm.Cpu.tickTimers()
	}
}

func (vm *VM) PeekNext() DisasmInfo {
	pc := vm.Cpu.pc
	op := vm.Memory.ReadU16(pc)
	asm := Disasm(op)
	return DisasmInfo{PC: pc, Op: op, Asm: asm}
}

func (vm *VM) Peek(n int) []DisasmInfo {
	copy := *vm
	results := make([]DisasmInfo, 0, n)

	for range n {
		pc := copy.Cpu.pc

		if int(pc)+1 >= len(copy.Memory.bytes) {
			break
		}

		op := copy.Memory.ReadU16(pc)
		asm := Disasm(op)

		results = append(results, DisasmInfo{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})

		copy.Step()
	}

	return results
}

func (vm *VM) DisasmRom() []DisasmInfo {
	start := ProgramStart
	end := ProgramStart + vm.RomSize
	results := make([]DisasmInfo, 0, vm.RomSize/OP_SIZE)

	for pc := uint16(start); pc < uint16(end); pc += OP_SIZE {
		op := vm.Memory.ReadU16(pc)
		asm := Disasm(op)

		results = append(results, DisasmInfo{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})
	}

	return results
}
