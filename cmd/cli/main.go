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
		args := strings.Fields(strings.TrimSpace(line))
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "help":
			printHelp()

		case "load":
			if len(args) < 2 {
				fmt.Println("Usage: load <rom>")
				continue
			}
			loadRom(emu, args[1])

		case "step":
			fmt.Println(emu.DisasmNext())
			fmt.Println()
			emu.Step()

		case "regs":
			fmt.Println(chip8.DebugRegisters(&emu.Cpu))
			fmt.Println()

		case "run":
			runCmd(emu, args)

		case "disasm", "d":
			disasmCmd(emu, args)

		case "draw":
			println(chip8.RenderASCII(&emu.Display))

		case "exit", "quit":
			return

		default:
			fmt.Println("Unknown command:", args[0])
		}
	}
}

func printHelp() {
	fmt.Println(`
Commands:
  help            Show a list of all supported commands.
  load <file>     Load ROM
  step            Execute one instruction
  run <steps>     Execute multiple steps
  regs            Print registers
  disasm <n>      Disassemble N instructions
  draw            Render display in ascii
  quit            Exit`)
	fmt.Println()
}

func loadRom(emu *chip8.Emu, path string) {
	rom, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	emu.LoadRom(rom)
	fmt.Printf("ROM loaded (%d bytes).\n", len(rom))
	fmt.Println()
}

func runCmd(emu *chip8.Emu, args []string) {
	steps := 10

	if len(args) >= 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid number:", args[1])
		}
		steps = n
	}

	for i := 0; i < steps; i++ {
		emu.Step()
	}

	fmt.Printf("Executed %d steps.\n", steps)
	fmt.Println()
}

func disasmCmd(emu *chip8.Emu, args []string) {
	n := 10

	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[1]); err == nil {
			n = v
		} else {
			fmt.Println("Invalid number:", args[0])
			return
		}
	}

	list := emu.DisasmN(n)
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}
