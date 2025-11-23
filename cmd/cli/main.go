package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mxmgorin/ch8go/chip8"
)

func main() {
	fmt.Println("CHIP-8 CLI. Type 'help' for commands.")

	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	emu := chip8.NewEmu()

	if *romPath != "" {
		loadRom(emu, *romPath)
	}

	for {
		fmt.Print("chip8> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "help":
			printHelp()

		case "load":
			if len(fields) < 2 {
				fmt.Println("Usage: load <rom>")
				continue
			}
			loadRom(emu, fields[1])

		case "step":
			emu.Step()
			fmt.Println(emu.Cpu.DebugRegisters())

		case "regs":
			fmt.Println(emu.Cpu.DebugRegisters())

		case "run":
			steps := 10

			if len(fields) >= 2 {
				n, err := strconv.Atoi(fields[1])
				if err != nil {
					fmt.Println("Invalid number:", fields[1])
				}
				steps = n
			}

			for i := 0; i < steps; i++ {
				emu.Step()
			}

			fmt.Printf("Executed %d steps.\n", steps)

		case "exit", "quit":
			return

		default:
			fmt.Println("Unknown command:", fields[0])
		}
	}
}

func printHelp() {
	fmt.Println(`
Commands:
  load <file>     Load ROM
  step            Execute one instruction
  run <steps>     Execute multiple steps
  regs            Print registers
  quit            Exit`)
}

func loadRom(emu *chip8.Emu, path string) {
	rom, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	emu.LoadRom(rom)
	fmt.Println("ROM loaded.")
}
