package core

const KEYS_COUNT = 16

type Keypad struct {
	keys [KEYS_COUNT]byte
}

func NewKeypad() *Keypad {
	return &Keypad{}
}

func (k *Keypad) Press(key byte) {
	if key < KEYS_COUNT {
		k.keys[key] = 1
	}
}

func (k *Keypad) Release(key byte) {
	if key < KEYS_COUNT {
		k.keys[key] = 0
	}
}

func (k *Keypad) IsPressed(key byte) bool {
	return key < KEYS_COUNT && k.keys[key] == 1
}

func (k *Keypad) Reset() {
	for i := range k.keys {
		k.keys[i] = 0
	}
}

func (k *Keypad) GetPressed() (byte, bool) {
	for i, v := range k.keys {
		if v == 1 {
			return byte(i), true
		}
	}
	return 0, false
}
