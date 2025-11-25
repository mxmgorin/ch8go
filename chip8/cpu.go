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
	rpl   [8]byte // schip extension
}

func NewCpu() Cpu {
	return Cpu{
		pc: 0x200,
	}
}

func (e *Cpu) Reset() {
	for i := range e.v {
		e.v[i] = 0
	}

	e.i = 0
	e.pc = 0x200
	e.sp = 0
	e.dt = 0
	e.st = 0

	for i := range e.stack {
		e.stack[i] = 0
	}
}

func (e *Cpu) tickTimers() {
	if e.dt > 0 {
		e.dt--
	}
	if e.st > 0 {
		e.st--
	}
}

func (e *Cpu) fetch(memory *Memory) uint16 {
	opcode := memory.ReadU16(e.pc)
	e.pc += 2

	return opcode
}

func (e *Cpu) execute(op uint16, memory *Memory, display *Display, keypad *Keypad) {
	switch op & 0xF000 {
	case 0x0000:
		switch op & 0x00FF {

		case 0xC0: // 00CN
			n := read_n(op)
			display.ScrollDown(n)

		case 0xE0: // 00E0 - CLS
			display.Clear()

		case 0xEE: // 00EE - RET
			e.ret()

		case 0xFB:
			display.ScrollRight4()

		case 0xFC:
			display.ScrollLeft4()

		case 0xFD:
			e.Reset()

		case 0xFE: // 00FE - lowres schip
			display.setResolution(false)

		case 0xFF: // 00FF - hires schip
			display.setResolution(true)

		default:
			// 0NNN - SYS addr (ignored)
		}
	case 0x1000: // JP addr
		e.jp(read_nnn(op))

	case 0x2000: // CALL addr
		e.call(read_nnn(op))

	case 0x3000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		e.skipNextIf(e.v[x] == nn)

	case 0x4000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		e.skipNextIf(e.v[x] != nn)

	case 0x5000:
		x := read_x(op)
		y := read_y(op)
		e.skipNextIf(e.v[x] == e.v[y])

	case 0x6000: // LD Vx, byte
		x := read_x(op)
		e.v[x] = read_nn(op)

	case 0x7000: // ADD Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		e.v[x] += nn

	case 0x8000:
		e.execute_8xyn(op)

	case 0x9000:
		x := read_x(op)
		y := read_y(op)
		e.skipNextIf(e.v[x] != e.v[y])

	case 0xA000: // LD I, addr
		e.i = read_nnn(op)

	case 0xB000: // JP V0, addr
		e.jp(read_nnn(op) + uint16(e.v[0]))

	case 0xC000: // RND Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		e.v[x] = byte(rand.Intn(256)) & nn

	case 0xD000: // DRW Vx, Vy, nibble
		vx := e.v[read_x(op)]
		vy := e.v[read_y(op)]
		n := uint16(read_n(op))
		var collisions int

		if n == 0 {
			// SCHIP 16x16 sprite (32 bytes = 16 pixels, 2 bytes per row)
			sprite := memory.ReadSprite(e.i, 32)
			collisions = display.DrawSprite16(vx, vy, sprite)
		} else {
			// Classic CHIP-8 8×N sprite
			sprite := memory.ReadSprite(e.i, uint16(n))
			collisions = display.DrawSprite(vx, vy, sprite)
		}
		if collisions > 0 {
			e.v[0xF] = 1
		} else {
			e.v[0xF] = 0
		}

	case 0xE000:
		switch op & 0x00FF {
		case 0x9E: // SKP Vx
			e.skipNextIf(keypad.IsPressed(e.v[read_x(op)]))

		case 0xA1: // SKNP Vx
			e.skipNextIf(!keypad.IsPressed(e.v[read_x(op)]))

		default: // ignored
		}

	case 0xF000:
		e.execute_fnnn(op, memory, keypad)

	default:
		// todo: handle others
	}
}

func (e *Cpu) execute_8xyn(op uint16) {
	x := read_x(op)
	y := read_y(op)

	switch op & 0x000F {
	case 0x0: // LD Vx, Vy
		e.v[x] = e.v[y]

	case 0x1: // OR Vx, Vy
		e.v[x] |= e.v[y]

	case 0x2: // AND Vx, Vy
		e.v[x] &= e.v[y]

	case 0x3: // XOR Vx, Vy
		e.v[x] ^= e.v[y]

	case 0x4: // ADD Vx, Vy (with carry)
		sum := uint16(e.v[x]) + uint16(e.v[y])
		carry := byte(sum >> 8)
		e.v[x] = byte(sum) // store result FIRST
		e.v[0xF] = carry

	case 0x5: // SUB Vx, Vy (Vx = Vx - Vy)
		vy := e.v[y]
		vx := e.v[x]
		borrow := byte(0)
		if vx >= vy {
			borrow = 1
		}

		e.v[x] = vx - vy // store result FIRST
		e.v[0xF] = borrow

	case 0x6: // SHR Vx {, Vy} – shifts Vx right by 1
		vy := e.v[y]
		flag := vy & 0x1
		e.v[x] = vy >> 1
		e.v[0xF] = flag

	case 0x7: // SUBN Vx, Vy (Vx = Vy - Vx)
		vy := e.v[y]
		vx := e.v[x]
		borrow := byte(0)
		if vy >= vx {
			borrow = 1
		}

		e.v[x] = vy - vx // store result FIRST
		e.v[0xF] = borrow

	case 0xE: // SHL Vx, Vy
		vy := e.v[y]
		flag := (vy >> 7) & 0x1
		e.v[x] = vy << 1
		e.v[0xF] = flag

	default:
		// Unknown 8XY* instruction
	}
}

func (e *Cpu) execute_fnnn(op uint16, memory *Memory, keypad *Keypad) {
	x := read_x(op)

	switch op & 0x00FF {
	case 0x07: // LD Vx, DT
		e.v[x] = e.dt

	case 0x0A: // LD Vx, K
		key, pressed := keypad.GetPressed()
		if pressed {
			e.v[x] = key
		} else {
			// Don't advance PC → repeat this opcode next cycle
			e.pc -= 2
		}

	case 0x15: // LD DT, Vx
		e.dt = e.v[x]

	case 0x18: // LD ST, Vx
		e.st = e.v[x]

	case 0x1E: // ADD I, Vx
		e.i += uint16(e.v[x])

	case 0x30: // FX30 - point I to 8x10 big digit sprite
		digit := e.v[x] & 0x0F
		e.i = uint16(BigFontStart) + uint16(digit)*16

	case 0x33:
		val := e.v[x]
		memory.Write(e.i+0, val/100)     // hundreds
		memory.Write(e.i+1, (val/10)%10) // tens
		memory.Write(e.i+2, val%10)      // ones

	case 0x55:
		for r := uint16(0); r <= uint16(x); r++ {
			memory.Write(e.i+r, e.v[r])
		}

	case 0x65:
		for r := uint16(0); r <= uint16(x); r++ {
			e.v[r] = memory.Read(e.i + r)
		}

	case 0x75: // FX75 - store V0..VX into RPL flags
		for i := byte(0); i <= byte(x) && i < 8; i++ {
			e.rpl[i] = e.v[i]

		}

	case 0x85: // FX85 - load V0..VX from RPL flags
		for i := byte(0); i <= byte(x) && i < 8; i++ {
			e.v[i] = e.rpl[i]
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

func (e *Cpu) push(val uint16) {
	e.stack[e.sp] = val
	e.sp += 1
}

func (e *Cpu) pop() uint16 {
	e.sp -= 1
	return e.stack[e.sp]
}

func (e *Cpu) ret() {
	e.pc = e.pop()
}

func (e *Cpu) call(addr uint16) {
	e.push(e.pc)
	e.pc = addr
}

func (e *Cpu) jp(addr uint16) {
	e.pc = addr
}

func (e *Cpu) skipNextIf(cond bool) {
	if cond {
		e.pc += 2
	}
}
