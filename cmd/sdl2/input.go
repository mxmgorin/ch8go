package main

import (
	"github.com/mxmgorin/ch8go/pkg/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

var keymap = map[sdl.Keycode]chip8.Key{
	sdl.K_1: chip8.Key1,
	sdl.K_2: chip8.Key2,
	sdl.K_3: chip8.Key3,
	sdl.K_4: chip8.KeyC,

	sdl.K_q: chip8.Key4,
	sdl.K_w: chip8.Key5,
	sdl.K_e: chip8.Key6,
	sdl.K_r: chip8.KeyD,

	sdl.K_a: chip8.Key7,
	sdl.K_s: chip8.Key8,
	sdl.K_d: chip8.Key9,
	sdl.K_f: chip8.KeyE,

	sdl.K_z: chip8.KeyA,
	sdl.K_x: chip8.Key0,
	sdl.K_c: chip8.KeyB,
	sdl.K_v: chip8.KeyF,
}

func handleKey(key sdl.Keycode, keypad *chip8.Keypad, down bool) {
	if k, ok := keymap[key]; ok {
		keypad.HandleKey(k, down)
	}
}
