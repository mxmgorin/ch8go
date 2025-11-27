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

type CLI struct {
	app *app.App
}

func newCLI() CLI {
	a, _ := app.NewApp(&ASCIIPainter{})
	return CLI{app: a}
}

type ASCIIPainter struct{}

func (p *ASCIIPainter) Init(w, h int) error {
	return nil
}

func (p *ASCIIPainter) Destroy() {}

func (p *ASCIIPainter) Paint(rgbaBuf []byte, sc app.Color, w, h int) {
	const (
		on  = "██"
		off = "░░"
	)

	out := strings.Builder{}
	out.Grow(h * w * 2)

	for y := range h {
		for x := range w {
			i := (y*w + x) * 4 // RGBA = 4 bytes per pixel
			r := rgbaBuf[i]
			g := rgbaBuf[i+1]
			b := rgbaBuf[i+2]

			if (r | g | b) != 0 {
				out.WriteString(on)
			} else {
				out.WriteString(off)
			}
		}

		out.WriteByte('\n')
	}

	ascii := out.String()
	println(ascii)
}

func (cli *CLI) run() {
	reader := bufio.NewReader(os.Stdin)

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
			cli.loadROM(args[1])

		case "step":
			cli.step(args)

		case "regs":
			cli.regs()

		case "peek":
			cli.peek(args)

		case "draw":
			cli.draw()

		case "dis":
			cli.dis()

		case "info":
			info := cli.app.ROMInfo()
			b, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(b))

		case "exit", "quit":
			return

		default:
			fmt.Println("Unknown command:", args[0])
		}
	}
}

func (cli *CLI) loadROM(path string) {
	len, err := cli.app.ReadROM(path)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("ROM loaded (%d bytes).\n", len)
	}

	fmt.Println()
}

func (cli *CLI) regs() {
	if cli.noROM() {
		return
	}

	fmt.Println(chip8.DebugRegisters(&cli.app.VM.CPU))
	fmt.Println()
}

func (cli *CLI) step(args []string) {
	if cli.noROM() {
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
		cli.app.VM.Step()
	}

	if steps > 1 {
		fmt.Printf("Executed %d steps.\n", steps)
	} else {
		fmt.Println(cli.app.VM.PeekNext())
	}

	fmt.Println()
}

func (cli *CLI) draw() {
	if cli.noROM() {
		return
	}

	// cli.app.Paint()
	println(chip8.RenderASCII(&cli.app.VM.Display))
	fmt.Println()
}

func (cli *CLI) peek(args []string) {
	if cli.noROM() {
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

	list := cli.app.VM.Peek(n)
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}

func (cli *CLI) dis() {
	if cli.noROM() {
		return
	}

	list := cli.app.VM.DisasmROM()
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}

func (cli *CLI) noROM() bool {
	if !cli.app.HasROM() {
		fmt.Println("No ROM. Use 'load <file>' first.")
		fmt.Println()
		return true
	}

	return false
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

func main() {
	fmt.Println("ch8go CLI. Type 'help' for commands.")

	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	flag.Parse()

	cli := newCLI()
	if *romPath != "" {
		cli.loadROM(*romPath)
	}

	cli.run()
}
