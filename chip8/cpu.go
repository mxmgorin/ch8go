package chip8

import (
	"math/rand"
)

const OP_SIZE = 2

type Cpu struct {
	v     [16]byte
	i     uint16
	pc    uint16
	sp    byte
	stack [16]uint16
	dt    byte
	st    byte
	hz    int
	timer float64
}

func NewCpu() Cpu {
	return Cpu{
		pc: 0x200,
		hz: 500,
	}
}

func (c *Cpu) Reset() {
	for i := range c.v {
		c.v[i] = 0
	}

	c.hz = 500
	c.i = 0
	c.pc = 0x200
	c.sp = 0
	c.dt = 0
	c.st = 0
	c.timer = 0

	for i := range c.stack {
		c.stack[i] = 0
	}
}

func (c *Cpu) tickTimers() {
	c.timer += 60.0 / float64(c.hz)

	for c.timer >= 1.0 { // second passed
		c.timer -= 1.0

		if c.dt > 0 {
			c.dt--
		}
		if c.st > 0 {
			c.st--
		}
	}
}

func (c *Cpu) fetch(memory *Memory) uint16 {
	opcode := memory.ReadU16(c.pc)
	c.pc += 2

	return opcode
}

func (c *Cpu) execute(op uint16, memory *Memory, display *Display, keypad *Keypad) {
	switch op & 0xF000 {
	case 0x0000:
		switch op & 0x00FF {
		case 0xE0: // 00E0 - CLS
			display.Clear()

		case 0xEE: // 00EE - RET
			c.ret()

		default:
			// 0NNN - SYS addr (ignored)
		}
	case 0x1000: // JP addr
		c.jp(read_nnn(op))

	case 0x2000: // CALL addr
		c.call(read_nnn(op))

	case 0x3000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		c.skipNextIf(c.v[x] == nn)

	case 0x4000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		c.skipNextIf(c.v[x] != nn)

	case 0x5000:
		x := read_x(op)
		y := read_y(op)
		c.skipNextIf(c.v[x] == c.v[y])

	case 0x6000: // LD Vx, byte
		x := read_x(op)
		c.v[x] = read_nn(op)

	case 0x7000: // ADD Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		c.v[x] += nn

	case 0x8000:
		c.execute_8xyn(op)

	case 0x9000:
		x := read_x(op)
		y := read_y(op)
		c.skipNextIf(c.v[x] != c.v[y])

	case 0xA000: // LD I, addr
		c.i = read_nnn(op)

	case 0xB000: // JP V0, addr
		c.jp(read_nnn(op) + uint16(c.v[0]))

	case 0xC000: // RND Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		c.v[x] = byte(rand.Intn(256)) & nn

	case 0xD000: // DRW Vx, Vy, nibble
		x := c.v[read_x(op)]
		y := c.v[read_y(op)]
		height := uint16(read_n(op))
		sprite := memory.ReadSprite(c.i, height)
		collision := display.DrawSprite(x, y, sprite)

		if collision {
			c.v[0xF] = 1
		} else {
			c.v[0xF] = 0
		}

	case 0xE000:
		switch op & 0x00FF {
		case 0x9E: // SKP Vx
			c.skipNextIf(keypad.IsPressed(c.v[read_x(op)]))

		case 0xA1: // SKNP Vx
			c.skipNextIf(!keypad.IsPressed(c.v[read_x(op)]))

		default: // ignored
		}

	case 0xF000:
		c.execute_fnnn(op, memory, keypad)

	default:
		// todo: handle others
	}
}

func (c *Cpu) execute_8xyn(op uint16) {
	x := read_x(op)
	y := read_y(op)

	switch op & 0x000F {
	case 0x0: // LD Vx, Vy
		c.v[x] = c.v[y]

	case 0x1: // OR Vx, Vy
		c.v[x] |= c.v[y]

	case 0x2: // AND Vx, Vy
		c.v[x] &= c.v[y]

	case 0x3: // XOR Vx, Vy
		c.v[x] ^= c.v[y]

	case 0x4: // ADD Vx, Vy (with carry)
		sum := uint16(c.v[x]) + uint16(c.v[y])
		carry := byte(sum >> 8)
		c.v[x] = byte(sum) // store result FIRST
		c.v[0xF] = carry

	case 0x5: // SUB Vx, Vy (Vx = Vx - Vy)
		vy := c.v[y]
		vx := c.v[x]
		borrow := byte(0)
		if vx >= vy {
			borrow = 1
		}

		c.v[x] = vx - vy // store result FIRST
		c.v[0xF] = borrow

	case 0x6: // SHR Vx {, Vy} – shifts Vx right by 1
		vy := c.v[y]
		flag := vy & 0x1
		c.v[x] = vy >> 1
		c.v[0xF] = flag

	case 0x7: // SUBN Vx, Vy (Vx = Vy - Vx)
		vy := c.v[y]
		vx := c.v[x]
		borrow := byte(0)
		if vy >= vx {
			borrow = 1
		}

		c.v[x] = vy - vx // store result FIRST
		c.v[0xF] = borrow

	case 0xE: // SHL Vx, Vy
		vy := c.v[y]
		flag := (vy >> 7) & 0x1
		c.v[x] = vy << 1
		c.v[0xF] = flag

	default:
		// Unknown 8XY* instruction
	}
}

func (c *Cpu) execute_fnnn(op uint16, memory *Memory, keypad *Keypad) {
	x := (op >> 8) & 0x0F

	switch op & 0x00FF {
	case 0x07: // LD Vx, DT
		c.v[x] = c.dt

	case 0x0A: // LD Vx, K
		key, pressed := keypad.GetPressed()
		if pressed {
			c.v[x] = key
		} else {
			// Don't advance PC → repeat this opcode next cycle
			c.pc -= 2
		}

	case 0x15: // LD DT, Vx
		c.dt = c.v[x]

	case 0x18: // LD ST, Vx
		c.st = c.v[x]

	case 0x1E: // ADD I, Vx
		c.i += uint16(c.v[x])

	case 0x33:
		val := c.v[x]
		memory.Write(c.i+0, val/100)     // hundreds
		memory.Write(c.i+1, (val/10)%10) // tens
		memory.Write(c.i+2, val%10)      // ones

	case 0x55:
		for r := uint16(0); r <= uint16(x); r++ {
			memory.Write(c.i+r, c.v[r])
		}

	case 0x65:
		for r := uint16(0); r <= uint16(x); r++ {
			c.v[r] = memory.Read(c.i + r)
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

func (c *Cpu) push(val uint16) {
	c.stack[c.sp] = val
	c.sp += 1
}

func (c *Cpu) pop() uint16 {
	c.sp -= 1
	return c.stack[c.sp]
}

func (c *Cpu) ret() {
	c.pc = c.pop()
}

func (c *Cpu) call(addr uint16) {
	c.push(c.pc)
	c.pc = addr
}

func (c *Cpu) jp(addr uint16) {
	c.pc = addr
}

func (c *Cpu) skipNextIf(cond bool) {
	if cond {
		c.pc += 2
	}
}
