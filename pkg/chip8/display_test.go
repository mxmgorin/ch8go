package chip8

import "testing"

func TestOpPlane(t *testing.T) {
	d := NewDisplay()

	d.opPlane(3) // enable planes 0 and 1
	if d.planeMask != 3 {
		t.Errorf("opPlane: planeMask = %d, want 3", d.planeMask)
	}
	if d.isPlaneDisabled(0) || d.isPlaneDisabled(1) {
		t.Error("planes 0 and 1 should be enabled")
	}
	if !d.isPlaneDisabled(2) {
		t.Error("plane 2 should be disabled")
	}
}
