//go:build js && wasm

package main

import (
	"log/slog"
)

var app App

func main() {
	slog.Info("ch8go WASM")
	app = newApp()
	app.run()
}
