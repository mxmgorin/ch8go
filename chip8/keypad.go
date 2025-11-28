package chip8

const KEYS_COUNT = 16

type Keypad struct {
	keys     [KEYS_COUNT]bool
	prevKeys [KEYS_COUNT]bool // previous state
}

func NewKeypad() Keypad {
	return Keypad{}
}

func (k *Keypad) HandleKey(key byte, pressed bool) {
	if key < KEYS_COUNT {
		k.prevKeys[key] = k.keys[key]
		k.keys[key] = pressed
	}
}

func (k *Keypad) Press(key byte) {
	if key < KEYS_COUNT {
		k.prevKeys[key] = k.keys[key]
		k.keys[key] = true
	}
}

func (k *Keypad) Release(key byte) {
	if key < KEYS_COUNT {
		k.prevKeys[key] = k.keys[key]
		k.keys[key] = false
	}
}

func (k *Keypad) IsPressed(key byte) bool {
	return key < KEYS_COUNT && k.keys[key]
}

func (k *Keypad) Reset() {
	for key := range k.keys {
		k.keys[key] = false
	}
}

func (k *Keypad) GetReleased() (key byte, ok bool) {
	for i := range k.keys {
		if k.prevKeys[i] && !k.keys[i] {
			return byte(i), true
		}
	}
	return 0, false
}

func (k *Keypad) Update() {
	copy(k.prevKeys[:], k.keys[:])
}
