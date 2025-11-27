package chip8

import "fmt"

type VM struct {
	CPU      CPU
	Memory   Memory
	Display  Display
	Keypad   Keypad
	RomSize  int
	tickrate int
}

func NewVM() *VM {
	return &VM{
		CPU:      NewCpu(QuirksSuperChip11),
		Memory:   NewMemory(),
		Display:  NewDisplay(),
		Keypad:   NewKeypad(),
		tickrate: 15,
	}
}

func (vm *VM) SetTickrate(r int)  { vm.tickrate = r }
func (vm *VM) SetQuirks(q Quirks) { vm.CPU.quirks = q }

func (vm *VM) LoadROM(bytes []byte) error {
	err := vm.Memory.Load(bytes)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}
	vm.Display.Clear()
	vm.Keypad.Reset()
	vm.CPU.Reset()
	vm.RomSize = len(bytes)
	vm.tickrate = 15
	vm.CPU.quirks = QuirksSuperChip11

	return nil
}

func (vm *VM) Step() {
	if !vm.Display.pendingVBlank || !vm.CPU.quirks.VBlankWait {
		opcode := vm.CPU.fetch(&vm.Memory)
		vm.CPU.execute(opcode, &vm.Memory, &vm.Display, &vm.Keypad)
	}
}

// Should run at 60 fps
func (vm *VM) RunFrame() bool {
	for range vm.tickrate {
		vm.Step()
	}

	vm.CPU.tickTimers()
	return vm.Display.poll()
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
