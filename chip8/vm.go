package chip8

import "fmt"

const defaultCPUHz = 700.0

var defaultQuirks = QuirksSChipModern

type VM struct {
	CPU        CPU
	Memory     Memory
	Display    Display
	Keypad     Keypad
	RomSize    int
	cpuHz      float64
	timerHz    float64
	cycleAccum float64
	timerAccum float64
}

func NewVM() *VM {
	return &VM{
		CPU:     NewCpu(defaultQuirks),
		Memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
		timerHz: 60.0,
		cpuHz:   defaultCPUHz,
	}
}

func (vm *VM) SetTickrate(tr int) { vm.cpuHz = float64(tr) * 60.0 }
func (vm *VM) SetQuirks(q Quirks) { vm.CPU.quirks = q }

func (vm *VM) LoadROM(bytes []byte) error {
	vm.Reset()

	if err := vm.Memory.Load(bytes); err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	return nil
}

func (vm *VM) Reset() {
	vm.Memory.Reset()
	vm.Display.Reset()
	vm.Keypad.Reset()
	vm.CPU.Reset()
	vm.cpuHz = defaultCPUHz
	vm.CPU.quirks = defaultQuirks
}

func (vm *VM) Step() {
	if !vm.Display.pendingVBlank || !vm.CPU.quirks.VBlankWait {
		opcode := vm.CPU.fetch(&vm.Memory)
		vm.CPU.execute(opcode, &vm.Memory, &vm.Display, &vm.Keypad)
	}
}

func (vm *VM) RunFrame(dt float64) bool {
	vm.cycleAccum += vm.cpuHz * dt

	for vm.cycleAccum >= 1 {
		vm.cycleAccum -= 1
		vm.Step()
	}

	vm.timerAccum += vm.timerHz * dt

	for vm.timerAccum >= 1 {
		vm.timerAccum -= 1
		vm.CPU.tickTimers()
	}

	vm.Keypad.Latch()

	return vm.Display.poll()
}

func (vm *VM) Buzzer() bool {
	return vm.CPU.st > 0
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
