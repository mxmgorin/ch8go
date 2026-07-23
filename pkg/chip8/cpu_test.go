package chip8

import "testing"

func TestOpRND(t *testing.T) {
	c := NewCpu(DefaultConf.Quirks)

	// A zero mask forces the result to 0 regardless of the random value.
	c.v[5] = 0xFF
	c.opRND(0xC500)
	if c.v[5] != 0 {
		t.Errorf("opRND mask 0x00: v5 = %#02x, want 0", c.v[5])
	}

	// A 0x0F mask must always leave the high nibble clear.
	for i := 0; i < 32; i++ {
		c.opRND(0xC50F)
		if c.v[5]&0xF0 != 0 {
			t.Fatalf("opRND mask 0x0F set high nibble: %#02x", c.v[5])
		}
	}
}

func TestOpF000(t *testing.T) {
	vm := NewVM()
	// F000 reads the following word as a 16-bit address into I.
	vm.Memory.bytes[0x200] = 0x12
	vm.Memory.bytes[0x201] = 0x34
	vm.CPU.pc = 0x200

	vm.CPU.Execute(0xF000, &vm.Memory, &vm.Display, &vm.Keypad, &vm.Audio)

	if vm.CPU.i != 0x1234 {
		t.Errorf("opF000: i = %#04x, want 0x1234", vm.CPU.i)
	}
}

func TestOp5XY2And5XY3(t *testing.T) {
	vm := NewVM()
	c := &vm.CPU
	c.i = 0x300

	// Ascending range (x < y): save V2..V3, then load them back.
	c.v[2], c.v[3] = 0xAA, 0xBB
	c.op5XY2(&vm.Memory, 0x5232)
	if vm.Memory.bytes[0x300] != 0xAA || vm.Memory.bytes[0x301] != 0xBB {
		t.Fatalf("op5XY2 save: mem = %#02x %#02x", vm.Memory.bytes[0x300], vm.Memory.bytes[0x301])
	}
	c.v[2], c.v[3] = 0, 0
	c.op5XY3(&vm.Memory, 0x5232)
	if c.v[2] != 0xAA || c.v[3] != 0xBB {
		t.Fatalf("op5XY3 load: v2=%#02x v3=%#02x", c.v[2], c.v[3])
	}

	// Descending range (x > y): exercises the reverse-iteration branch.
	c.v[4], c.v[5] = 0x11, 0x22
	c.op5XY2(&vm.Memory, 0x5542)
	if vm.Memory.bytes[0x300] != 0x22 || vm.Memory.bytes[0x301] != 0x11 {
		t.Fatalf("op5XY2 reverse: mem = %#02x %#02x", vm.Memory.bytes[0x300], vm.Memory.bytes[0x301])
	}
	c.v[4], c.v[5] = 0, 0
	c.op5XY3(&vm.Memory, 0x5542)
	if c.v[5] != 0x22 || c.v[4] != 0x11 {
		t.Fatalf("op5XY3 reverse: v4=%#02x v5=%#02x", c.v[4], c.v[5])
	}
}

func TestVMTickrate(t *testing.T) {
	vm := NewVM()
	vm.SetTickrate(30)
	if got := vm.Tickrate(); got != 30 {
		t.Errorf("Tickrate() = %d, want 30", got)
	}
}
