package core

type Keypad struct {
	Keys [16]bool
}

func NewKeypad() *Keypad {
	return &Keypad{}
}

func (k *Keypad) Press(key byte) {
	if key < 16 {
		k.Keys[key] = true
	}
}

func (k *Keypad) Release(key byte) {
	if key < 16 {
		k.Keys[key] = false
	}
}

func (k *Keypad) IsPressed(key byte) bool {
	return key < 16 && k.Keys[key]
}

func (k *Keypad) Clear() {
	for i := range k.Keys {
		k.Keys[i] = false
	}
}
