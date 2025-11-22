package core

type Cpu struct {
	v     [16]byte
	i     uint16
	pc    uint16
	sp    byte
	stack [16]uint16
}

func NewCpu() *Cpu {
	return &Cpu{
		pc: 0x200,
	}
}

func (cpu *Cpu) Step(memory *Memory, display *Display, keypad *Keypad) {
	hi := memory.Read(cpu.pc)
	lo := memory.Read(cpu.pc + 1)
	opcode := uint16(hi)<<8 | uint16(lo)
	cpu.pc += 2

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
