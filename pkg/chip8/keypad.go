package chip8

type Key uint8

const (
	Key0 Key = iota // 0
	Key1            // 1
	Key2            // 2
	Key3            // 3
	Key4            // 4
	Key5            // 5
	Key6            // 6
	Key7            // 7
	Key8            // 8
	Key9            // 9
	KeyA            // A
	KeyB            // B
	KeyC            // C
	KeyD            // D
	KeyE            // E
	KeyF            // F
	KeyCount
)

type Keypad struct {
	keys     [KeyCount]bool
	prevKeys [KeyCount]bool // previous state
}

func NewKeypad() Keypad {
	return Keypad{}
}

func (k *Keypad) HandleKey(key Key, pressed bool) {
	if key < KeyCount {
		k.prevKeys[key] = k.keys[key]
		k.keys[key] = pressed
	}
}

func (k *Keypad) Press(key Key) {
	if key < KeyCount {
		k.prevKeys[key] = k.keys[key]
		k.keys[key] = true
	}
}

func (k *Keypad) Release(key Key) {
	if key < KeyCount {
		k.prevKeys[key] = k.keys[key]
		k.keys[key] = false
	}
}

func (k *Keypad) IsPressed(key Key) bool {
	return key < KeyCount && k.keys[key]
}

func (k *Keypad) Reset() {
	for key := range k.keys {
		k.keys[key] = false
		k.prevKeys[key] = false
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

func (k *Keypad) Latch() {
	copy(k.prevKeys[:], k.keys[:])
}
