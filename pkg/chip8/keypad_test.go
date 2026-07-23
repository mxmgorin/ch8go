package chip8

import "testing"

func TestKeypadPressRelease(t *testing.T) {
	k := NewKeypad()

	if k.IsPressed(Key5) {
		t.Error("fresh keypad should report no keys pressed")
	}

	k.Press(Key5)
	if !k.IsPressed(Key5) {
		t.Error("Key5 should be pressed after Press")
	}

	k.Release(Key5)
	if k.IsPressed(Key5) {
		t.Error("Key5 should not be pressed after Release")
	}
}

func TestKeypadHandleKey(t *testing.T) {
	k := NewKeypad()

	k.HandleKey(Key3, true)
	if !k.IsPressed(Key3) {
		t.Error("HandleKey(Key3, true) should press Key3")
	}

	k.HandleKey(Key3, false)
	if k.IsPressed(Key3) {
		t.Error("HandleKey(Key3, false) should release Key3")
	}
}

func TestKeypadGetReleased(t *testing.T) {
	k := NewKeypad()

	if _, ok := k.GetReleased(); ok {
		t.Error("no key should be reported released on a fresh keypad")
	}

	// Press, latch the state, then release: the transition is now detectable.
	k.Press(KeyA)
	k.Latch()
	k.Release(KeyA)

	key, ok := k.GetReleased()
	if !ok || key != byte(KeyA) {
		t.Errorf("GetReleased() = (%d, %v), want (%d, true)", key, ok, byte(KeyA))
	}
}

func TestKeypadReset(t *testing.T) {
	k := NewKeypad()
	k.Press(Key1)
	k.Press(Key2)

	k.Reset()

	if k.IsPressed(Key1) || k.IsPressed(Key2) {
		t.Error("Reset should clear all pressed keys")
	}
	if _, ok := k.GetReleased(); ok {
		t.Error("Reset should clear previous-key state")
	}
}

func TestKeypadOutOfRange(t *testing.T) {
	k := NewKeypad()

	// Keys at or beyond KeyCount must be ignored without panicking.
	k.Press(KeyCount)
	k.HandleKey(KeyCount, true)
	if k.IsPressed(KeyCount) {
		t.Error("out-of-range key must never register as pressed")
	}
}
