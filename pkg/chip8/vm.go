package chip8

import "fmt"

type FrameState struct {
	Dirty bool
	Beep  bool
}

type VM struct {
	CPU        CPU
	Memory     Memory
	Display    Display
	Keypad     Keypad
	Audio      Audio
	romSize    int
	cpuHz      float64
	timerHz    float64
	cycleAccum float64
	timerAccum float64
}

func NewVM() *VM {
	return &VM{
		CPU:     NewCpu(DefaultConf.Quirks),
		Memory:  NewMemory(),
		Display: NewDisplay(),
		Keypad:  NewKeypad(),
		Audio:   NewAudio(),
		timerHz: 60.0,
		cpuHz:   DefaultConf.CPUHz(),
	}
}

func (vm *VM) SetConf(conf PlatformConf) {
	vm.SetQuirks(conf.Quirks)
	vm.SetTickrate(conf.Tickrate)
	vm.Audio.SetMode(conf.AudioMode)
}

func (vm *VM) Tickrate() int      { return int(vm.cpuHz / 60.0) }
func (vm *VM) SetTickrate(tr int) { vm.cpuHz = float64(tr) * 60.0 }
func (vm *VM) SetQuirks(q Quirks) { vm.CPU.Quirks = q }

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
	vm.Audio.Reset()
	vm.cpuHz = DefaultConf.CPUHz()
	vm.CPU.Quirks = DefaultConf.Quirks
}

func (vm *VM) Step() {
	if !vm.Display.pendingVBlank || !vm.CPU.Quirks.WaitVBlank {
		opcode := vm.CPU.fetch(&vm.Memory)
		vm.CPU.execute(opcode, &vm.Memory, &vm.Display, &vm.Keypad, &vm.Audio)
	}
}

func (vm *VM) RunFrame(dt float64) FrameState {
	state := FrameState{}
	vm.cycleAccum += vm.cpuHz * dt

	for vm.cycleAccum >= 1 {
		vm.cycleAccum -= 1
		vm.Step()
	}

	vm.timerAccum += vm.timerHz * dt

	for vm.timerAccum >= 1 {
		vm.timerAccum -= 1
		vm.CPU.tickTimer()
		state.Beep = vm.Audio.TickTimer()
	}

	vm.Keypad.Latch()
	state.Dirty = vm.Display.poll()

	return state
}

func (vm *VM) PeekNext() Instruction {
	pc := vm.CPU.pc
	op := vm.Memory.ReadU16(pc)
	asm := Disasm(op)
	return Instruction{PC: pc, Op: op, Asm: asm}
}

func (vm *VM) Peek(n int) []Instruction {
	copy := *vm
	results := make([]Instruction, 0, n)

	for range n {
		pc := copy.CPU.pc

		if int(pc)+1 >= len(copy.Memory.bytes) {
			break
		}

		op := copy.Memory.ReadU16(pc)
		asm := Disasm(op)

		results = append(results, Instruction{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})

		copy.Step()
	}

	return results
}

func (vm *VM) DisasmROM() []Instruction {
	start := ProgramStart
	end := ProgramStart + vm.romSize
	results := make([]Instruction, 0, vm.romSize/opSize)

	for pc := uint16(start); pc < uint16(end); pc += opSize {
		op := vm.Memory.ReadU16(pc)
		asm := Disasm(op)

		results = append(results, Instruction{
			PC:  pc,
			Op:  op,
			Asm: asm,
		})
	}

	return results
}
