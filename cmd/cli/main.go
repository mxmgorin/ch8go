package main

import (
	"flag"
	"fmt"
)

func main() {
	fmt.Println("ch8go a. Type 'help' for commands.")

	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	flag.Parse()

	a := newApp()
	if *romPath != "" {
		a.cmdLoad([]string{"load", *romPath})
	}

	a.run()
}
