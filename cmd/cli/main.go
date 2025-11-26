package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mxmgorin/ch8go/app"
	"github.com/mxmgorin/ch8go/chip8"
)

var romHash string

func main() {
	fmt.Println("ch8go CLI. Type 'help' for commands.")

	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	app := app.NewApp()

	if *romPath != "" {
		loadRom(app, *romPath)
	}

	for {
		fmt.Print("ch8go> ")
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
			loadRom(app, args[1])

		case "step":
			step(&app.Emu, args)

		case "regs":
			regs(&app.Emu)

		case "peek":
			peek(&app.Emu, args)

		case "draw":
			draw(&app.Emu)

		case "dis":
			dis(&app.Emu)

		case "info":
			info := app.ROMInfo()
			b, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(b))

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
  help            Show all supported commands
  load <file>     Load a ROM into memory
  step <n>        Execute 1 or N instructions
  peek <n>        Disassemble 1 or N instructions starting from PC
  regs            Show registers
  dis             Disassemble the loaded ROM
  draw            Render the current display buffer in ASCII
  info            Show metadata about a ROM
  quit            Exit`)
	fmt.Println()
}

func loadRom(app *app.App, path string) {
	len := app.LoadRom(path)
	fmt.Printf("ROM loaded (%d bytes).\n", len)
	fmt.Println()
}

func regs(emu *chip8.Emu) {
	if noRom(emu) {
		return
	}

	fmt.Println(chip8.DebugRegisters(&emu.Cpu))
	fmt.Println()
}

func step(emu *chip8.Emu, args []string) {
	if noRom(emu) {
		return
	}

	steps := 1

	if len(args) >= 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid number:", args[1])
		}
		steps = n
	}

	for range steps {
		emu.Step()
	}

	if steps > 1 {
		fmt.Printf("Executed %d steps.\n", steps)
	} else {
		fmt.Println(emu.PeekNext())
	}

	fmt.Println()
}

func draw(emu *chip8.Emu) {
	if noRom(emu) {
		return
	}

	println(chip8.RenderASCII(&emu.Display))
	fmt.Println()
}

func peek(emu *chip8.Emu, args []string) {
	if noRom(emu) {
		return
	}

	n := 10

	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[1]); err == nil {
			n = v
		} else {
			fmt.Println("Invalid number:", args[0])
			return
		}
	}

	list := emu.Peek(n)
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}

func dis(e *chip8.Emu) {
	if noRom(e) {
		return
	}

	list := e.DisasmRom()
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}

func noRom(e *chip8.Emu) bool {
	if e.Status == chip8.StatusNoRom {
		fmt.Println("No ROM. Use 'load <file>' first.")
		fmt.Println()
		return true
	}

	return false
}
