//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

type Input struct {
	keymap map[string]chip8.Key
}

func newInput() Input {
	i := Input{
		keymap: map[string]chip8.Key{
			"1": chip8.Key1, "2": chip8.Key2, "3": chip8.Key3, "4": chip8.KeyC,
			"q": chip8.Key4, "w": chip8.Key5, "e": chip8.Key6, "r": chip8.KeyD,
			"a": chip8.Key7, "s": chip8.Key8, "d": chip8.Key9, "f": chip8.KeyE,
			"z": chip8.KeyA, "x": chip8.Key0, "c": chip8.KeyB, "v": chip8.KeyF,
			"ArrowUp": chip8.Key5, "ArrowDown": chip8.Key8, "ArrowLeft": chip8.Key7, "ArrowRight": chip8.Key9,
			" ": chip8.Key6,
		},
	}

	win := js.Global().Get("window")
	win.Call("addEventListener", "keydown", js.FuncOf(i.onKeyDown))
	win.Call("addEventListener", "keyup", js.FuncOf(i.onKeyUp))

	return i
}

type KeyEvent struct {
	Key     chip8.Key
	Pressed bool
}

func (i *Input) onKeyDown(this js.Value, args []js.Value) any {
	return i.onKey(args[0], true)
}

func (i *Input) onKeyUp(this js.Value, args []js.Value) any {
	return i.onKey(args[0], false)
}

func (i *Input) onKey(event js.Value, pressed bool) any {
	// Detect if user is typing into an input, textarea, or contenteditable
	target := event.Get("target")
	nodeName := target.Get("nodeName").String()
	isContentEditable := target.Get("isContentEditable").Truthy()

	if nodeName == "INPUT" || nodeName == "TEXTAREA" || isContentEditable {
		// Don't block keypresses on input fields
		return nil
	}

	key := event.Get("key").String()
	if k, ok := i.keymap[key]; ok {
		app.keyChan <- KeyEvent{Key: k, Pressed: pressed}
		event.Call("preventDefault")
	}
	return nil
}
