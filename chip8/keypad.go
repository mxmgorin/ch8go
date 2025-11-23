package chip8

const KEYS_COUNT = 16

type Keypad struct {
	Keys [KEYS_COUNT]bool
}

func NewKeypad() Keypad {
	return Keypad{}
}

func (k *Keypad) Press(key byte) {
	if key < KEYS_COUNT {
		k.Keys[key] = true
	}
}

func (k *Keypad) Release(key byte) {
	if key < KEYS_COUNT {
		k.Keys[key] = false
	}
}

func (k *Keypad) IsPressed(key byte) bool {
	return key < KEYS_COUNT && k.Keys[key]
}

func (k *Keypad) Reset() {
	for i := range k.Keys {
		k.Keys[i] = false
	}
}

func (k *Keypad) GetPressed() (byte, bool) {
	for i, v := range k.Keys {
		if v {
			return byte(i), true
		}
	}
	return 0, false
}
