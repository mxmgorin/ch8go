package chip8

import (
	"fmt"
	"strings"
)

type DisasmInfo struct {
	PC  uint16
	Op  uint16
	Asm string
}

func (d DisasmInfo) String() string {
	return fmt.Sprintf("%04X: %04X  %s", d.PC, d.Op, d.Asm)
}

func DebugRegisters(cpu *Cpu) string {
	return fmt.Sprintf("PC=%04X I=%04X V=%v", cpu.pc, cpu.i, cpu.v)
}

func RenderASCII(d *Display) string {
	const (
		on  = "█"
		off = "░"
	)

	out := strings.Builder{}
	out.Grow(DisplayHeight * DisplayWidth * 2)

	for y := range DisplayHeight {
		for x := range DisplayWidth {
			if d.Pixels[y*DisplayWidth+x] != 0 {
				out.WriteString(on)
			} else {
				out.WriteString(off)
			}
		}
		out.WriteByte('\n')
	}

	return out.String()
}

func Disasm(op uint16) string {
	x := read_x(op)
	y := read_y(op)
	n := read_n(op)
	nn := read_nn(op)
	nnn := read_nnn(op)

	switch op & 0xF000 {
	case 0x0000:
		switch nn {
		case 0xE0:
			return "CLS"
		case 0xEE:
			return "RET"
		}
		return fmt.Sprintf("SYS %03X", nnn)
	case 0x1000:
		return fmt.Sprintf("JP  %03X", nnn)
	case 0x2000:
		return fmt.Sprintf("CALL %03X", nnn)
	case 0x3000:
		return fmt.Sprintf("SE  V%X, %02X", x, nn)
	case 0x4000:
		return fmt.Sprintf("SNE V%X, %02X", x, nn)
	case 0x5000:
		return fmt.Sprintf("SE  V%X, V%X", x, y)
	case 0x6000:
		return fmt.Sprintf("LD  V%X, %02X", x, nn)
	case 0x7000:
		return fmt.Sprintf("ADD V%X, %02X", x, nn)
	case 0x8000:
		switch n {
		case 0x0:
			return fmt.Sprintf("LD  V%X, V%X", x, y)
		case 0x1:
			return fmt.Sprintf("OR  V%X, V%X", x, y)
		case 0x2:
			return fmt.Sprintf("AND V%X, V%X", x, y)
		case 0x3:
			return fmt.Sprintf("XOR V%X, V%X", x, y)
		case 0x4:
			return fmt.Sprintf("ADD V%X, V%X", x, y)
		case 0x5:
			return fmt.Sprintf("SUB V%X, V%X", x, y)
		case 0x6:
			return fmt.Sprintf("SHR V%X", x)
		case 0x7:
			return fmt.Sprintf("SUBN V%X, V%X", x, y)
		case 0xE:
			return fmt.Sprintf("SHL V%X", x)
		}
	case 0x9000:
		return fmt.Sprintf("SNE V%X, V%X", x, y)
	case 0xA000:
		return fmt.Sprintf("LD  I, %03X", nnn)
	case 0xB000:
		return fmt.Sprintf("JP  V0, %03X", nnn)
	case 0xC000:
		return fmt.Sprintf("RND V%X, %02X", x, nn)
	case 0xD000:
		return fmt.Sprintf("DRW V%X, V%X, %X", x, y, n)
	case 0xE000:
		switch nn {
		case 0x9E:
			return fmt.Sprintf("SKP V%X", x)
		case 0xA1:
			return fmt.Sprintf("SKNP V%X", x)
		}
	case 0xF000:
		switch nn {
		case 0x07:
			return fmt.Sprintf("LD  V%X, DT", x)
		case 0x0A:
			return fmt.Sprintf("LD  V%X, K", x)
		case 0x15:
			return fmt.Sprintf("LD  DT, V%X", x)
		case 0x18:
			return fmt.Sprintf("LD  ST, V%X", x)
		case 0x1E:
			return fmt.Sprintf("ADD I, V%X", x)
		case 0x29:
			return fmt.Sprintf("LD  F, V%X", x)
		case 0x33:
			return fmt.Sprintf("BCD V%X", x)
		case 0x55:
			return fmt.Sprintf("LD  [I], V0-V%X", x)
		case 0x65:
			return fmt.Sprintf("LD  V0-V%X, [I]", x)
		}
	}

	return fmt.Sprintf(".DW %04X", op)

}
