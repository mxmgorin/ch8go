package chip8

import (
	"strings"
	"testing"
)

func TestDisasm(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
	}{
		{0x00E0, "CLS"},
		{0x00EE, "RET"},
		{0x0123, "SYS 123"},
		{0x1234, "JP  234"},
		{0x2345, "CALL 345"},
		{0x3A42, "SE  VA, 42"},
		{0x4B10, "SNE VB, 10"},
		{0x5AB0, "SE  VA, VB"},
		{0x6C42, "LD  VC, 42"},
		{0x7D01, "ADD VD, 01"},
		{0x8AB0, "LD  VA, VB"},
		{0x8AB1, "OR  VA, VB"},
		{0x8AB2, "AND VA, VB"},
		{0x8AB3, "XOR VA, VB"},
		{0x8AB4, "ADD VA, VB"},
		{0x8AB5, "SUB VA, VB"},
		{0x8AB6, "SHR VA"},
		{0x8AB7, "SUBN VA, VB"},
		{0x8ABE, "SHL VA"},
		{0x9AB0, "SNE VA, VB"},
		{0xA123, "LD  I, 123"},
		{0xB234, "JP  V0, 234"},
		{0xCA42, "RND VA, 42"},
		{0xDAB5, "DRW VA, VB, 5"},
		{0xEA9E, "SKP VA"},
		{0xEAA1, "SKNP VA"},
		{0xFA07, "LD  VA, DT"},
		{0xFA0A, "LD  VA, K"},
		{0xFA15, "LD  DT, VA"},
		{0xFA18, "LD  ST, VA"},
		{0xFA1E, "ADD I, VA"},
		{0xFA29, "LD  F, VA"},
		{0xFA33, "BCD VA"},
		{0xFA55, "LD  [I], V0-VA"},
		{0xFA65, "LD  V0-VA, [I]"},
		// Unknown opcodes fall through to the raw-word form.
		{0x8ABF, ".DW 8ABF"},
		{0xEA00, ".DW EA00"},
		{0xFA00, ".DW FA00"},
	}

	for _, tt := range tests {
		if got := Disasm(tt.op); got != tt.want {
			t.Errorf("Disasm(%#04x) = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestInstructionString(t *testing.T) {
	got := Instruction{PC: 0x0200, Op: 0x00E0, Asm: "CLS"}.String()
	if want := "0200: 00E0  CLS"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestRegistersString(t *testing.T) {
	cpu := NewCpu(DefaultConf.Quirks)
	cpu.pc = 0x0200
	cpu.i = 0x0123
	cpu.v[0] = 0xAB

	got := RegistersString(&cpu)
	for _, want := range []string{"PC=0200", "I=0123", "V=[171 "} {
		if !strings.Contains(got, want) {
			t.Errorf("RegistersString() = %q, missing %q", got, want)
		}
	}
}

func TestRenderASCII(t *testing.T) {
	d := NewDisplay()
	size := d.Size()

	empty := RenderASCII(&d)
	if lines := strings.Count(empty, "\n"); lines != size.Height {
		t.Errorf("RenderASCII() newlines = %d, want %d", lines, size.Height)
	}
	if strings.Contains(empty, "██") {
		t.Error("RenderASCII() of a cleared display should contain no on-pixels")
	}
	if !strings.Contains(empty, "░░") {
		t.Error("RenderASCII() of a cleared display should contain off-pixels")
	}

	d.Planes[0][0] = 1
	if !strings.Contains(RenderASCII(&d), "██") {
		t.Error("RenderASCII() should render a set pixel as on")
	}
}

func TestVMPeekAndDisasmROM(t *testing.T) {
	vm := NewVM()
	// LD V0, 01 ; LD V1, 02
	rom := []byte{0x60, 0x01, 0x61, 0x02}
	if err := vm.LoadROM(rom); err != nil {
		t.Fatalf("LoadROM() error = %v", err)
	}

	if next := vm.PeekNext(); next.PC != 0x0200 || next.Op != 0x6001 || next.Asm != "LD  V0, 01" {
		t.Errorf("PeekNext() = %+v, want {0200 6001 LD  V0, 01}", next)
	}

	want := []Instruction{
		{PC: 0x0200, Op: 0x6001, Asm: "LD  V0, 01"},
		{PC: 0x0202, Op: 0x6102, Asm: "LD  V1, 02"},
	}

	peek := vm.Peek(2)
	if len(peek) != len(want) {
		t.Fatalf("Peek(2) len = %d, want %d", len(peek), len(want))
	}
	for i, w := range want {
		if peek[i] != w {
			t.Errorf("Peek()[%d] = %+v, want %+v", i, peek[i], w)
		}
	}

	dis := vm.DisasmROM()
	if len(dis) != len(want) {
		t.Fatalf("DisasmROM() len = %d, want %d", len(dis), len(want))
	}
	for i, w := range want {
		if dis[i] != w {
			t.Errorf("DisasmROM()[%d] = %+v, want %+v", i, dis[i], w)
		}
	}
}
