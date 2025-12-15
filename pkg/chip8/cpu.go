package chip8

import (
	"math/rand"
)

const opSize = 2

// CPU represents the CHIP-8 processor state.
type CPU struct {
	v      [16]byte
	i      uint16
	pc     uint16
	sp     byte
	stack  [256]uint16 // original is 16 but modern games require deeper stack
	dt     byte
	flags  [16]byte // xochip ext, schip has 8
	Quirks Quirks
}

func NewCpu(quirks Quirks) CPU {
	return CPU{
		pc:     0x200,
		Quirks: quirks,
	}
}

func (c *CPU) Reset() {
	for i := range c.v {
		c.v[i] = 0
	}

	c.i = 0
	c.pc = 0x200
	c.sp = 0
	c.dt = 0

	for i := range c.stack {
		c.stack[i] = 0
	}
}

func (c *CPU) tickTimer() {
	if c.dt > 0 {
		c.dt--
	}
}

func (c *CPU) fetch(memory *Memory) uint16 {
	opcode := memory.ReadU16(c.pc)
	c.pc += 2

	return opcode
}

// Execute executes a single instruction.
//
// It fetches the opcode at PC, decodes it, executes it,
// and updates timers as specified by the specification.
func (c *CPU) Execute(op uint16, memory *Memory, display *Display, keypad *Keypad, audio *Audio) {
	switch op & 0xF000 {
	case 0x0000:
		switch op & 0x00FF {
		case 0xE0: // 00E0 - CLS
			display.opClear()

		case 0xEE: // 00EE - RET
			c.ret()

		case 0xFB:
			display.ScrollRight4(c.Quirks.ScaleScroll)

		case 0xFC:
			display.ScrollLeft4(c.Quirks.ScaleScroll)

		case 0xFD:
			c.Reset()

		case 0xFE: // 00FE - lowres schip
			display.opRes(false)

		case 0xFF: // 00FF - hires schip
			display.opRes(true)

		default:
			// Special: 00CN (00C0 – 00CF)
			switch op & 0x00F0 {
			case 0x00C0: // 00CN, schip
				n := read_n(op)
				display.ScrollDown(n, c.Quirks.ScaleScroll)
			case 0x00D0: // 00DN , xochip
				n := read_n(op)
				display.ScrollUp(int(n))
			}
		}
	case 0x1000: // JP addr
		c.jp(read_nnn(op))

	case 0x2000: // CALL addr
		c.call(read_nnn(op))

	case 0x3000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		c.skipNextIf(memory, c.v[x] == nn)

	case 0x4000: // SE Vx, byte
		x := read_x(op)
		nn := read_nn(op)
		c.skipNextIf(memory, c.v[x] != nn)

	case 0x5000:
		switch op & 0x000F {
		case 0x0: // 5XY0
			x := read_x(op)
			y := read_y(op)
			c.skipNextIf(memory, c.v[x] == c.v[y])

		case 0x2: // xochip
			c.op5XY2(memory, op)

		case 0x3: // xochip
			c.op5XY3(memory, op)
		}

	case 0x6000: // LD Vx, byte
		x := read_x(op)
		c.v[x] = read_nn(op)

	case 0x7000: // ADD Vx, byte
		c.opADD(op)

	case 0x8000:
		c.op8XYN(op)

	case 0x9000:
		x := read_x(op)
		y := read_y(op)
		c.skipNextIf(memory, c.v[x] != c.v[y])

	case 0xA000: // LD I, addr
		c.i = read_nnn(op)

	case 0xB000: // JP V0, addr
		c.opJP(op)

	case 0xC000: // RND Vx, byte
		c.opRND(op)

	case 0xD000: // DRW Vx, Vy, nibble
		c.opDRAW(op, memory, display)

	case 0xE000:
		switch op & 0x00FF {
		case 0x9E: // SKP Vx
			c.skipNextIf(memory, keypad.IsPressed(Key(c.v[read_x(op)])))

		case 0xA1: // SKNP Vx
			c.skipNextIf(memory, !keypad.IsPressed(Key(c.v[read_x(op)])))

		default: // ignored
		}

	case 0xF000:
		c.opFNNN(op, display, memory, keypad, audio)

	default:
	}
}

func (c *CPU) opJP(op uint16) {
	var v byte
	if c.Quirks.Jump {
		x := read_x(op)
		v = c.v[x]
	} else {
		v = c.v[0]
	}
	c.jp(read_nnn(op) + uint16(v))
}

func (c *CPU) opADD(op uint16) {
	x := read_x(op)
	nn := read_nn(op)
	c.v[x] += nn
}

func (c *CPU) opRND(op uint16) {
	x := read_x(op)
	nn := read_nn(op)
	c.v[x] = byte(rand.Intn(256)) & nn
}

func (c *CPU) opDRAW(op uint16, memory *Memory, display *Display) {
	vx := c.v[read_x(op)]
	vy := c.v[read_y(op)]
	n := uint16(read_n(op))
	var collisions int

	if n == 0 {
		// SCHIP 16x16 sprite (32 bytes = 16 pixels, 2 bytes per row)
		const height = 16
		const bytesPerRow = 2
		endAddr := height * bytesPerRow * uint16(display.planesLen())
		sprite := memory.ReadSprite(c.i, endAddr)
		collisions = display.DrawSprite(vx, vy, sprite, 16, height, bytesPerRow, c.Quirks.Wrap)
	} else {
		// Classic CHIP-8 8×N sprite
		endAddr := n * uint16(display.planesLen())
		sprite := memory.ReadSprite(c.i, endAddr)
		collisions = display.DrawSprite(vx, vy, sprite, 8, int(n), 1, c.Quirks.Wrap)
	}
	if collisions > 0 {
		c.v[0xF] = 1
	} else {
		c.v[0xF] = 0
	}
}

func (c *CPU) op8XYN(op uint16) {
	x := read_x(op)
	y := read_y(op)

	switch op & 0x000F {
	case 0x0: // LD Vx, Vy
		c.v[x] = c.v[y]

	case 0x1: // OR Vx, Vy
		c.v[x] |= c.v[y]
		if c.Quirks.ResetFlag {
			c.v[0xF] = 0
		}

	case 0x2: // AND Vx, Vy
		c.v[x] &= c.v[y]
		if c.Quirks.ResetFlag {
			c.v[0xF] = 0
		}

	case 0x3: // XOR Vx, Vy
		c.v[x] ^= c.v[y]
		if c.Quirks.ResetFlag {
			c.v[0xF] = 0
		}

	case 0x4: // ADD Vx, Vy (with carry)
		sum := uint16(c.v[x]) + uint16(c.v[y])
		carry := byte(sum >> 8)
		c.v[x] = byte(sum) // store result FIRST
		c.v[0xF] = carry

	case 0x5: // SUB Vx, Vy (Vx = Vx - Vy)
		c.opSUB(x, y)

	case 0x6: // SHR Vx {, Vy} – shifts Vx right by 1
		c.opSHR(x, y)

	case 0x7: // SUBN Vx, Vy (Vx = Vy - Vx)
		c.opSUBN(x, y)

	case 0xE: // SHL Vx, Vy
		c.opSHL(x, y)

	default:
	}
}

func (c *CPU) opSUB(x, y uint16) {
	vy := c.v[y]
	vx := c.v[x]
	borrow := byte(0)
	if vx >= vy {
		borrow = 1
	}

	c.v[x] = vx - vy // store result FIRST
	c.v[0xF] = borrow
}

func (c *CPU) opSHR(x, y uint16) {
	in := y
	if c.Quirks.Shift {
		in = x
	}
	v := c.v[in]
	c.v[x] = v >> 1
	c.v[0xF] = v & 0x1 // LSB
}

func (c *CPU) opSUBN(x, y uint16) {
	vy := c.v[y]
	vx := c.v[x]
	borrow := byte(0)
	if vy >= vx {
		borrow = 1
	}

	c.v[x] = vy - vx // store result FIRST
	c.v[0xF] = borrow
}

func (c *CPU) opSHL(x, y uint16) {
	in := y
	if c.Quirks.Shift {
		in = x
	}
	v := c.v[in]
	c.v[x] = v << 1
	c.v[0xF] = (v >> 7) & 0x1 // MSB
}

// XOCHIP. i := long NNNN (0xF000, 0xNNNN) load i with a 16-bit address.
func (c *CPU) opF000(mem *Memory) {
	// Read the next 16-bit word as the address
	addr := c.fetch(mem)
	c.i = addr
}

func (c *CPU) opFNNN(op uint16, display *Display, memory *Memory, keypad *Keypad, audio *Audio) {
	if op == 0xF000 {
		c.opF000(memory)
		return
	}

	x := read_x(op)

	switch op & 0x00FF {
	case 0x01:
		display.opPlane(read_x(op))

	case 0x02: // audio
		audio.opPattern(memory, c.i)

	case 0x07: // LD Vx, DT
		c.v[x] = c.dt

	case 0x0A: // LD Vx, K
		c.opF0A(x, keypad)

	case 0x15: // LD DT, Vx
		c.dt = c.v[x]

	case 0x18: // LD ST, Vx
		audio.st = c.v[x]

	case 0x1E: // ADD I, Vx
		c.i += uint16(c.v[x])

	case 0x29: // Fx29 - small (4x5) digit
		digit := c.v[x] & 0x0F
		c.i = fontAddr + uint16(digit)*5

	case 0x30: // Fx30 - big (8x10) digit
		digit := c.v[x] & 0x0F
		c.i = bigFontAddr + uint16(digit)*10

	case 0x33:
		c.opF33(x, memory)

	case 0x3A: // pitch
		audio.opPitch(c.v[x])

	case 0x55:
		for r := uint16(0); r <= uint16(x); r++ {
			memory.Write(c.i+r, c.v[r])
		}
		c.Quirks.opMem(c, x)

	case 0x65:
		for r := uint16(0); r <= uint16(x); r++ {
			c.v[r] = memory.Read(c.i + r)
		}
		c.Quirks.opMem(c, x)

	case 0x75: // FX75 - store V0..VX into flags
		for i := byte(0); i <= byte(x); i++ {
			c.flags[i] = c.v[i]
		}

	case 0x85: // FX85 - load V0..VX from flags
		for i := byte(0); i <= byte(x); i++ {
			c.v[i] = c.flags[i]
		}

	default:
	}
}

func (c *CPU) opF0A(x uint16, keypad *Keypad) {
	key, pressed := keypad.GetReleased()
	if pressed {
		c.v[x] = key
	} else {
		c.pc -= 2 // repeat instruction
	}
}

func (c *CPU) opF33(x uint16, memory *Memory) {
	val := c.v[x]
	memory.Write(c.i+0, val/100)
	memory.Write(c.i+1, (val/10)%10)
	memory.Write(c.i+2, val%10)
}

// Save Vx..Vy to memory at I
func (c *CPU) op5XY2(mem *Memory, op uint16) {
	x := read_x(op)
	y := read_y(op)

	dist := int(x) - int(y)
	if dist < 0 {
		dist = -dist
	}

	if x < y {
		for z := 0; z <= dist; z++ {
			mem.bytes[c.i+uint16(z)] = c.v[int(x)+z]
		}
	} else {
		for z := 0; z <= dist; z++ {
			mem.bytes[c.i+uint16(z)] = c.v[int(x)-z]
		}
	}
}

// 5XY3: Load Vx..Vy from memory at I
func (c *CPU) op5XY3(mem *Memory, op uint16) {
	x := read_x(op)
	y := read_y(op)

	dist := int(x) - int(y)
	if dist < 0 {
		dist = -dist
	}

	if x < y {
		for z := 0; z <= dist; z++ {
			c.v[int(x)+z] = mem.bytes[c.i+uint16(z)]
		}
	} else {
		for z := 0; z <= dist; z++ {
			c.v[int(x)-z] = mem.bytes[c.i+uint16(z)]
		}
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

func (c *CPU) push(val uint16) {
	c.stack[c.sp] = val
	c.sp += 1
}

func (c *CPU) pop() uint16 {
	c.sp -= 1
	return c.stack[c.sp]
}

func (c *CPU) ret() {
	c.pc = c.pop()
}

func (c *CPU) call(addr uint16) {
	c.push(c.pc)
	c.pc = addr
}

func (c *CPU) jp(addr uint16) {
	c.pc = addr
}

func (c *CPU) skipNextIf(mem *Memory, cond bool) {
	if cond {
		if c.fetch(mem) == 0xF000 { // 4 bytes opcode
			c.pc += 2
		}
	}
}
