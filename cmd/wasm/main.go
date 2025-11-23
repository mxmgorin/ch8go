//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/mxmgorin/ch8go/chip8"
)

const (
	W = chip8.DisplayWidth
	H = chip8.DisplayHeight
)

var (
	ctx       js.Value
	imageData js.Value
	rgbaBuf   []byte
	emu       *chip8.Emu
	lastTime  time.Time
	romLoaded bool

	loopFunc js.Func
)

func main() {
	fmt.Println("ch8go WASM")
	emu = chip8.NewEmu()

	// Setup canvas
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "chip8-canvas")
	canvas.Set("width", W)
	canvas.Set("height", H)
	ctx = canvas.Call("getContext", "2d")
	imageData = ctx.Call("createImageData", W, H)
	rgbaBuf = make([]byte, W*H*4)

	// Export ROM loader
	js.Global().Set("chip8_loadROM", js.FuncOf(loadROM))

	// Handle keyboard
	win := js.Global().Get("window")
	win.Call("addEventListener", "keydown", js.FuncOf(onKeyDown))
	win.Call("addEventListener", "keyup", js.FuncOf(onKeyUp))

	// Animation loop (must persist function or GC will kill it)
	loopFunc = js.FuncOf(loop)
	lastTime = time.Now()
	js.Global().Call("requestAnimationFrame", loopFunc)

	// Keep WASM alive
	select {}
}

func loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	emu.LoadRom(buf)
	romLoaded = true
	fmt.Println("Rom loaded")
	return nil
}

func onKeyDown(this js.Value, args []js.Value) any {
	key := args[0].Get("key").String()
	if k, ok := keymap[key]; ok {
		emu.Keypad.Press(k)
		args[0].Call("preventDefault")
	}
	return nil
}

func onKeyUp(this js.Value, args []js.Value) any {
	key := args[0].Get("key").String()
	if k, ok := keymap[key]; ok {
		emu.Keypad.Release(k)
		args[0].Call("preventDefault")
	}
	return nil
}

var keymap = map[string]byte{
	"1": 0x1, "2": 0x2, "3": 0x3, "4": 0xC,
	"q": 0x4, "w": 0x5, "e": 0x6, "r": 0xD,
	"a": 0x7, "s": 0x8, "d": 0x9, "f": 0xE,
	"z": 0xA, "x": 0x0, "c": 0xB, "v": 0xF,
}

func loop(this js.Value, args []js.Value) any {
	if !romLoaded {
		// Don't run CPU until ROM exists
		js.Global().Call("requestAnimationFrame", loopFunc)
		return nil
	}
	now := time.Now()
	dt := now.Sub(lastTime).Seconds()
	lastTime = now
	// Run CPU approx. 500Hz
	steps := max(int(500*dt), 1)

	for range steps {
		emu.Step()
	}

	drawScreen()

	// Schedule next frame
	js.Global().Call("requestAnimationFrame", loopFunc)
	return nil
}

func drawScreen() {
	pixels := emu.Display.Pixels

	for i := range pixels {
		v := byte(0)
		if pixels[i] != 0 {
			v = 255
		}
		idx := i * 4
		rgbaBuf[idx] = v
		rgbaBuf[idx+1] = v
		rgbaBuf[idx+2] = v
		rgbaBuf[idx+3] = 255
	}

	js.CopyBytesToJS(imageData.Get("data"), rgbaBuf)
	ctx.Call("putImageData", imageData, 0, 0)
}
