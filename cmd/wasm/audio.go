//go:build js && wasm

package main

import (
	"syscall/js"
	"unsafe"
)

type Audio struct {
	buf []float32
}

func newAudio() Audio {
	a := Audio{
		buf: make([]float32, 0),
	}

	js.Global().Set("fillAudio", js.FuncOf(a.output))
	js.Global().Set("startAudio", js.FuncOf(a.start))

	return a
}

func (a *Audio) start(this js.Value, args []js.Value) any {
	size := args[0].Int()
	a.buf = make([]float32, size)
	return nil
}

func (a *Audio) output(this js.Value, args []js.Value) any {
	out := args[0] // JS Float32Array
	freq := args[1].Float()
	app.emu.VM.Audio.Output(a.buf, freq)

	outBuffer := js.Global().Get("Uint8Array").New(
		out.Get("buffer"),
		out.Get("byteOffset"),
		out.Get("byteLength"),
	)
	bufPointer := unsafe.Pointer(&a.buf[0])
	byteBuf := unsafe.Slice((*byte)(bufPointer), len(a.buf)*4)
	js.CopyBytesToJS(outBuffer, byteBuf)

	return nil
}
