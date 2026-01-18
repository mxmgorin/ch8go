package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mxmgorin/ch8go/pkg/chip8"
)

var keymap = map[ebiten.Key]chip8.Key{
	ebiten.Key1: chip8.Key1,
	ebiten.Key2: chip8.Key2,
	ebiten.Key3: chip8.Key3,
	ebiten.Key4: chip8.KeyC,

	ebiten.KeyQ: chip8.Key4,
	ebiten.KeyW: chip8.Key5,
	ebiten.KeyE: chip8.Key6,
	ebiten.KeyR: chip8.KeyD,

	ebiten.KeyA: chip8.Key7,
	ebiten.KeyS: chip8.Key8,
	ebiten.KeyD: chip8.Key9,
	ebiten.KeyF: chip8.KeyE,

	ebiten.KeyZ: chip8.KeyA,
	ebiten.KeyX: chip8.Key0,
	ebiten.KeyC: chip8.KeyB,
	ebiten.KeyV: chip8.KeyF,
}

func handleKeys(a *App) {
	for k, v := range keymap {
		if ebiten.IsKeyPressed(k) {
			a.VM.Keypad.Press(v)
		} else {
			a.VM.Keypad.Release(v)
		}
	}
}
