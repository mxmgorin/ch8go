package chip8

import "testing"

func TestRunBreakpoint(t *testing.T) {
	vm := NewVM()
	// LD V0,01 ; LD V1,02 ; LD V2,03 ; JP 0206 (self-loop)
	rom := []byte{0x60, 0x01, 0x61, 0x02, 0x62, 0x03, 0x12, 0x06}
	if err := vm.LoadROM(rom); err != nil {
		t.Fatalf("LoadROM() error = %v", err)
	}

	res := vm.Run(map[uint16]bool{0x0204: true}, nil, 100)
	if res.Reason != StopBreakpoint {
		t.Fatalf("Reason = %v, want StopBreakpoint", res.Reason)
	}
	if res.Steps != 2 {
		t.Errorf("Steps = %d, want 2", res.Steps)
	}
	if pc := vm.PeekNext().PC; pc != 0x0204 {
		t.Errorf("PC = %04X, want 0204", pc)
	}
}

func TestRunCap(t *testing.T) {
	vm := NewVM()
	rom := []byte{0x12, 0x00} // JP 0200 — infinite loop, breakpoint never reached
	if err := vm.LoadROM(rom); err != nil {
		t.Fatalf("LoadROM() error = %v", err)
	}

	res := vm.Run(map[uint16]bool{0x0400: true}, nil, 50)
	if res.Reason != StopMaxSteps {
		t.Errorf("Reason = %v, want StopMaxSteps", res.Reason)
	}
	if res.Steps != 50 {
		t.Errorf("Steps = %d, want 50 (cap)", res.Steps)
	}
}

func TestRunWatch(t *testing.T) {
	vm := NewVM()
	// LD V0,05 ; JP 0202 (self-loop) — V0 changes on the first step
	rom := []byte{0x60, 0x05, 0x12, 0x02}
	if err := vm.LoadROM(rom); err != nil {
		t.Fatalf("LoadROM() error = %v", err)
	}

	res := vm.Run(nil, []Watch{{Reg: true, Addr: 0}}, 100)
	if res.Reason != StopWatch {
		t.Fatalf("Reason = %v, want StopWatch", res.Reason)
	}
	if res.Steps != 1 {
		t.Errorf("Steps = %d, want 1", res.Steps)
	}
	if res.WatchOld != 0x00 || res.WatchNew != 0x05 {
		t.Errorf("watch change = %02X -> %02X, want 00 -> 05", res.WatchOld, res.WatchNew)
	}
	if res.WatchDesc != "V0" {
		t.Errorf("WatchDesc = %q, want V0", res.WatchDesc)
	}
}
