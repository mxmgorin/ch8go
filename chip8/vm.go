package chip8

import "fmt"

type VM struct {
	CPU     CPU
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
		CPU:     NewCpu(QuirksSuperChip11),
		Memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
		cpuHz:   600.0,
		fps:     60.0,
	}
}

func (vm *VM) LoadROM(bytes []byte) error {
	err := vm.Memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}
	vm.Display.Clear()
	vm.Keypad.Reset()
	vm.CPU.Reset()
	vm.RomSize = len(bytes)

	return nil
}

func (vm *VM) Step() {
	if !vm.Display.pendingVBlank || !vm.CPU.quirks.VBlankWait {
		opcode := vm.CPU.fetch(&vm.Memory)
		vm.CPU.execute(opcode, &vm.Memory, &vm.Display, &vm.Keypad)
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
		vm.CPU.tickTimers()
	}
}

func (vm *VM) PeekNext() DisasmInfo {
	pc := vm.CPU.pc
	op := vm.Memory.ReadU16(pc)
	asm := Disasm(op)
	return DisasmInfo{PC: pc, Op: op, Asm: asm}
}

func (vm *VM) Peek(n int) []DisasmInfo {
	copy := *vm
	results := make([]DisasmInfo, 0, n)

	for range n {
		pc := copy.CPU.pc

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

func (vm *VM) DisasmROM() []DisasmInfo {
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
