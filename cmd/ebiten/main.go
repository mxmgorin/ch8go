package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/mxmgorin/ch8go/pkg/host"
)

func main() {
	slog.Info("ch8go ebiten")

	fs := flag.NewFlagSet("ch8go", flag.ExitOnError)
	opts, err := host.ParseOptions(fs, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if err := opts.ValidateROMPath(); err != nil {
		log.Fatal(err)
	}

	app, err := newApp(opts.Scale)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := app.ReadROM(opts.ROMPath); err != nil {
		log.Fatal(err)
	}

	if err := app.run(); err != nil {
		log.Fatal(err)
	}
}
