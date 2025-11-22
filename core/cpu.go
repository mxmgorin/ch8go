package core

type Cpu struct {
	v     [16]byte
	i     uint16
	pc    uint16
	sp    byte
	stack [16]uint16
	dt    byte
	st    byte
}

func NewCpu() *Cpu {
	return &Cpu{
		pc: 0x200,
	}
}

func (cpu *Cpu) TickTimers() {
	if cpu.dt > 0 {
		cpu.dt--
	}
	if cpu.st > 0 {
		cpu.st--
	}
}

func (cpu *Cpu) Step(memory *Memory, display *Display, keypad *Keypad) {
	opcode := cpu.fetch(memory)
	cpu.execute(opcode)
}

func (cpu *Cpu) push(val uint16) {
	cpu.stack[cpu.sp] = val
	cpu.sp += 1
}

func (cpu *Cpu) pop() uint16 {
	cpu.sp -= 1
	return cpu.stack[cpu.sp]
}

func (cpu *Cpu) fetch(memory *Memory) uint16 {
	hi := memory.Read(cpu.pc)
	lo := memory.Read(cpu.pc + 1)
	cpu.pc += 2
	opcode := uint16(hi)<<8 | uint16(lo)

	return opcode
}

func (cpu *Cpu) execute(opcode uint16) {
	switch opcode & 0xF000 {
	case 0x1000: // JP addr
		cpu.pc = opcode & 0x0FFF

	case 0x6000: // LD Vx, byte
		x := (opcode >> 8) & 0x0F
		cpu.v[x] = byte(opcode & 0x00FF)

	case 0xA000: // LD I, addr
		cpu.i = opcode & 0x0FFF

	default:
		// todo: handle others
	}
}
