package chip8

import (
	"fmt"
	"time"
)

const (
	// Timer decrements at 60 Hz independent of CPU frequency.
	TimerHz = 60
)

type FrameState struct {
	Dirty bool
	Beep  bool
}

// VM represents a complete CHIP-8 virtual machine instance.
//
// It aggregates all core subsystems required for execution, including
// CPU, memory, display, input, and audio. A VM instance owns its internal
// timing state and is responsible for coordinating CPU execution and
// timer updates.
type VM struct {
	CPU        CPU
	Memory     Memory
	Display    Display
	Keypad     Keypad
	Audio      Audio
	romSize    int
	cpuHz      float64
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

	vm.romSize = len(bytes)

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
		vm.CPU.Execute(opcode, &vm.Memory, &vm.Display, &vm.Keypad, &vm.Audio)
	}
}

func (vm *VM) RunFrame(frameDelta time.Duration) FrameState {
	state := FrameState{}
	dt := frameDelta.Seconds()
	vm.cycleAccum += vm.cpuHz * dt

	for vm.cycleAccum >= 1 {
		vm.cycleAccum -= 1
		vm.Step()
	}

	vm.timerAccum += TimerHz * dt

	for vm.timerAccum >= 1 {
		vm.timerAccum -= 1
		vm.CPU.tickTimer()
		state.Beep = vm.Audio.TickTimer()
	}

	vm.Keypad.Latch()
	state.Dirty = vm.Display.poll()

	return state
}

// Poll clears a pending VBlank wait and reports whether the display changed.
// Exposed for headless hosts (e.g. the CLI debugger) that step outside RunFrame.
func (vm *VM) Poll() bool {
	return vm.Display.poll()
}

// StopReason explains why Run stopped.
type StopReason int

const (
	StopMaxSteps StopReason = iota
	StopBreakpoint
	StopWatch
)

// Watch is a debugger watchpoint on a V register or a memory address.
type Watch struct {
	Reg  bool   // true: V register (Addr is the index 0-15); false: memory address
	Addr uint16
}

func (w Watch) String() string {
	if w.Reg {
		return fmt.Sprintf("V%X", w.Addr)
	}
	return fmt.Sprintf("mem[%04X]", w.Addr)
}

func (w Watch) read(vm *VM) byte {
	if w.Reg {
		return vm.CPU.v[w.Addr]
	}
	return vm.Memory.Read(w.Addr)
}

// RunResult reports the outcome of Run.
type RunResult struct {
	Steps     int
	Reason    StopReason
	WatchDesc string // set when Reason == StopWatch
	WatchOld  byte
	WatchNew  byte
}

// Run steps until a breakpoint PC is reached, a watched value changes, or
// maxSteps is exhausted. The instruction at a breakpoint is NOT executed.
// Timers, the keypad latch, and the display advance once per simulated frame so
// ROMs waiting on the delay timer, key input, or VBlank make progress.
func (vm *VM) Run(bps map[uint16]bool, watches []Watch, maxSteps int) RunResult {
	tr := vm.Tickrate() // CPU cycles per 60Hz timer tick
	sinceTimer := 0

	prev := make([]byte, len(watches))
	for i, w := range watches {
		prev[i] = w.read(vm)
	}

	for steps := 1; steps <= maxSteps; steps++ {
		vm.Step()

		if sinceTimer++; tr <= 0 || sinceTimer >= tr {
			sinceTimer = 0
			vm.CPU.tickTimer()
			vm.Audio.TickTimer()
			vm.Keypad.Latch()
			vm.Display.poll()
		}

		for i, w := range watches {
			if cur := w.read(vm); cur != prev[i] {
				return RunResult{Steps: steps, Reason: StopWatch, WatchDesc: w.String(), WatchOld: prev[i], WatchNew: cur}
			}
		}

		if bps[vm.CPU.pc] {
			return RunResult{Steps: steps, Reason: StopBreakpoint}
		}
	}

	return RunResult{Steps: maxSteps, Reason: StopMaxSteps}
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
