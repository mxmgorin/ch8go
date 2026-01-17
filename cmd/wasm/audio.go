//go:build js && wasm

package main

import (
	"syscall/js"
	"unsafe"
)

var audioBuf = make([]float32, 0)

func startAudio(this js.Value, args []js.Value) any {
	size := args[0].Int()
	audioBuf = make([]float32, size)
	return nil
}

func outputAudio(this js.Value, args []js.Value) any {
	out := args[0] // JS Float32Array
	freq := args[1].Float()
	app.emu.VM.Audio.Output(audioBuf, freq)

	outBuffer := js.Global().Get("Uint8Array").New(
		out.Get("buffer"),
		out.Get("byteOffset"),
		out.Get("byteLength"),
	)
	bufPointer := unsafe.Pointer(&audioBuf[0])
	byteBuf := unsafe.Slice((*byte)(bufPointer), len(audioBuf)*4)
	js.CopyBytesToJS(outBuffer, byteBuf)

	return nil
}
