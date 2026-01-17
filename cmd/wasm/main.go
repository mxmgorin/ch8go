//go:build js && wasm

package main

import (
	"log/slog"
)

func main() {
	slog.Info("ch8go WASM")
	app := newApp()
	app.run()
}
