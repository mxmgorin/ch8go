package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mxmgorin/ch8go/core"
)

func main() {
	romPath := flag.String("rom", "", "Path to CHIP-8 ROM")
	hz := flag.Int("hz", 500, "Cycles per second")
	flag.Parse()

	if *romPath == "" {
		log.Fatal("You must provide a ROM: --rom path/to/file.ch8")
	}

	romBytes, err := os.ReadFile(*romPath)
	if err != nil {
		log.Fatalf("Failed to read ROM: %v", err)
	}
	chip := core.NewChip8()

	if err := chip.LoadRom(romBytes); err != nil {
		log.Fatalf("Failed to load ROM: %v", err)
	}

	fmt.Println("CHIP-8 CLI Emulator")
	fmt.Println("-------------------")
	fmt.Printf("ROM: %s\n", *romPath)
	fmt.Printf("Speed: %d Hz\n", *hz)
	fmt.Println("Press CTRL+C to stop.")

	cycleDelay := time.Second / time.Duration(*hz)

	for {
		start := time.Now()
		chip.Step()

		elapsed := time.Since(start)
		if elapsed < cycleDelay {
			time.Sleep(cycleDelay - elapsed)
		}
	}
}
