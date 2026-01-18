//go:build !js

package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/mxmgorin/ch8go/pkg/host"
)

func main() {
	slog.Info("ch8go SDL2")

	fs := flag.NewFlagSet("ch8go", flag.ExitOnError)
	opts, err := host.ParseOptions(fs, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if err := opts.ValidateROMPath(); err != nil {
		log.Fatal(err)
	}

	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	defer app.Quit()

	if _, err := app.ReadROM(opts.ROMPath); err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
