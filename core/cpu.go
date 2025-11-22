package core

import "math/rand"

const OP_SIZE = 2

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
	cpu.execute(opcode, memory, display, keypad)
}

func (cpu *Cpu) fetch(memory *Memory) uint16 {
	hi := memory.Read(cpu.pc)
	lo := memory.Read(cpu.pc + 1)
	cpu.pc += 2
	opcode := uint16(hi)<<8 | uint16(lo)

	return opcode
}

func (cpu *Cpu) execute(op uint16, memory *Memory, display *Display, keypad *Keypad) {
	switch op & 0xF000 {
	case 0x0000:
		switch op & 0x00FF {
		case 0xE0: // 00E0 - CLS
			display.Clear()

		case 0xEE: // 00EE - RET
			cpu.ret()

		default:
			// 0NNN - SYS addr (ignored)
		}
	case 0x1000: // JP addr
		cpu.jp(read_nnn(op))

	case 0x2000: // CALL addr
		cpu.call(read_nnn(op))

	case 0x3000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		cpu.skipNextIf(cpu.v[x] == nn)

	case 0x4000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		cpu.skipNextIf(cpu.v[x] != nn)

	case 0x5000:
		x := read_x(op)
		y := read_y(op)
		cpu.skipNextIf(cpu.v[x] == cpu.v[y])

	case 0x6000: // LD Vx, byte
		x := read_x(op)
		cpu.v[x] = read_nn(op)

	case 0x7000: // ADD Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		cpu.v[x] += nn

	case 0x8000:
		cpu.execute_8xyn(op)

	case 0x9000:
		x := read_x(op)
		y := read_y(op)
		cpu.skipNextIf(cpu.v[x] != cpu.v[y])

	case 0xA000: // LD I, addr
		cpu.i = read_nnn(op)

	case 0xB000: // JP V0, addr
		cpu.jp(read_nnn(op) + uint16(cpu.v[0]))

	case 0xC000: // RND Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		cpu.v[x] = byte(rand.Intn(256)) & nn

	case 0xD000: // DRW Vx, Vy, nibble
		x := cpu.v[read_x(op)]
		y := cpu.v[read_y(op)]
		height := uint16(read_n(op))
		sprite := memory.ReadSprite(cpu.i, height)
		collision := display.DrawSprite(x, y, sprite)

		if collision {
			cpu.v[0xF] = 1
		} else {
			cpu.v[0xF] = 0
		}

	case 0xE000:
		switch op & 0x00FF {
		case 0x9E: // SKP Vx
			cpu.skipNextIf(keypad.IsPressed(cpu.v[read_x(op)]))

		case 0xA1: // SKNP Vx
			cpu.skipNextIf(!keypad.IsPressed(cpu.v[read_x(op)]))

		default: // ignored
		}

	case 0xF000:
		cpu.execute_fnnn(op, memory, keypad)

	default:
		// todo: handle others
	}
}

func (cpu *Cpu) execute_8xyn(op uint16) {
	x := read_x(op)
	y := read_y(op)

	switch op & 0x000F {
	case 0x0: // LD Vx, Vy
		cpu.v[x] = cpu.v[y]

	case 0x1: // OR Vx, Vy
		cpu.v[x] |= cpu.v[y]

	case 0x2: // AND Vx, Vy
		cpu.v[x] &= cpu.v[y]

	case 0x3: // XOR Vx, Vy
		cpu.v[x] ^= cpu.v[y]

	case 0x4: // ADD Vx, Vy (with carry)
		sum := uint16(cpu.v[x]) + uint16(cpu.v[y])
		cpu.v[0xF] = 0
		if sum > 0xFF {
			cpu.v[0xF] = 1
		}
		cpu.v[x] = byte(sum)

	case 0x5: // SUB Vx, Vy (Vx = Vx - Vy)
		cpu.v[0xF] = 0
		if cpu.v[x] > cpu.v[y] {
			cpu.v[0xF] = 1
		}
		cpu.v[x] = cpu.v[x] - cpu.v[y]

	case 0x6: // SHR Vx {, Vy} – shifts Vx right by 1
		// VF = least significant bit *before* shift
		cpu.v[0xF] = cpu.v[x] & 0x1
		cpu.v[x] >>= 1

	case 0x7: // SUBN Vx, Vy (Vx = Vy - Vx)
		cpu.v[0xF] = 0
		if cpu.v[y] > cpu.v[x] {
			cpu.v[0xF] = 1
		}
		cpu.v[x] = cpu.v[y] - cpu.v[x]

	case 0xE: // SHL Vx {, Vy} – shifts Vx left by 1
		cpu.v[0xF] = (cpu.v[x] >> 7) & 0x1
		cpu.v[x] <<= 1

	default:
		// Unknown 8XY* instruction
	}
}

func (cpu *Cpu) execute_fnnn(op uint16, memory *Memory, keypad *Keypad) {
	x := (op >> 8) & 0x0F

	switch op & 0x00FF {
	case 0x07: // LD Vx, DT
		cpu.v[x] = cpu.dt

	case 0x0A: // LD Vx, K
		key, pressed := keypad.GetPressed()
		if pressed {
			cpu.v[x] = key
		} else {
			// Don't advance PC → repeat this opcode next cycle
			cpu.pc -= 2
		}

	case 0x15: // LD DT, Vx
		cpu.dt = cpu.v[x]

	case 0x18: // LD ST, Vx
		cpu.st = cpu.v[x]

	case 0x1E: // ADD I, Vx
		cpu.i += uint16(cpu.v[x])

	case 0x33:
		val := cpu.v[x]
		memory.Write(cpu.i+0, val/100)     // hundreds
		memory.Write(cpu.i+1, (val/10)%10) // tens
		memory.Write(cpu.i+2, val%10)      // ones

	case 0x55:
		for r := uint16(0); r <= uint16(x); r++ {
			memory.Write(cpu.i+r, cpu.v[r])
		}

	case 0x65:
		for r := uint16(0); r <= uint16(x); r++ {
			cpu.v[r] = memory.Read(cpu.i + r)
		}

	default: // ignored
	}
}

func read_x(op uint16) uint16 {
	return (op >> 8) & 0x0F
}

func read_y(op uint16) uint16 {
	return (op >> 4) & 0x0F
}

func read_n(op uint16) byte {
	return byte(op & 0x000F)
}

func read_nn(op uint16) byte {
	return byte(op & 0x00FF)
}

func read_nnn(op uint16) uint16 {
	return op & 0x0FFF
}

func (cpu *Cpu) push(val uint16) {
	cpu.stack[cpu.sp] = val
	cpu.sp += 1
}

func (cpu *Cpu) pop() uint16 {
	cpu.sp -= 1
	return cpu.stack[cpu.sp]
}

func (cpu *Cpu) ret() {
	cpu.pc = cpu.pop()
}

func (cpu *Cpu) call(addr uint16) {
	cpu.push(cpu.pc)
	cpu.pc = addr
}

func (cpu *Cpu) jp(addr uint16) {
	cpu.pc = addr
}

func (cpu *Cpu) skipNextIf(cond bool) {
	if cond {
		cpu.pc += 2
	}
}
