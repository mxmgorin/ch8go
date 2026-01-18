package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mxmgorin/ch8go/pkg/chip8"
	"github.com/mxmgorin/ch8go/pkg/host"
)

type App struct {
	emu     *host.Emu
	painter ASCIIPainter
}

func newApp() App {
	a, _ := host.NewEmu()
	return App{emu: a, painter: ASCIIPainter{}}
}

func (a *App) run() {
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

		cmd := args[0]

		if cmdFn, ok := cmds[cmd]; ok {
			if err := cmdFn(a, args); err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println("error:", err)
			}
		} else {
			fmt.Println("Unknown command:", args[0])
		}
	}
}

func (a *App) cmdInfo() {
	info := a.emu.ROMMeta()
	b, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(b))
}

func (a *App) cmdLoad(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: load <rom>")
		return
	}

	path := args[1]
	len, err := a.emu.ReadROM(path)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("ROM loaded (%d bytes).\n", len)
	}

	fmt.Println()
}

func (a *App) cmdRegs() {
	if a.loaded() {
		return
	}

	fmt.Println(chip8.RegistersString(&a.emu.VM.CPU))
	fmt.Println()
}

func (a *App) cmdStep(args []string) {
	if a.loaded() {
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
		a.emu.VM.Step()
	}

	if steps > 1 {
		fmt.Printf("Executed %d steps.\n", steps)
	} else {
		fmt.Println(a.emu.VM.PeekNext())
	}

	fmt.Println()
}

func (a *App) cmdDraw() {
	if a.loaded() {
		return
	}

	//a.painter.Paint()
	println(chip8.RenderASCII(&a.emu.VM.Display))
	fmt.Println()
}

func (a *App) cmdPeek(args []string) {
	if a.loaded() {
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

	list := a.emu.VM.Peek(n)
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}

func (a *App) cmdDis() {
	if a.loaded() {
		return
	}

	list := a.emu.VM.DisasmROM()
	for _, info := range list {
		fmt.Println(info)
	}

	fmt.Println()
}

func (a *App) loaded() bool {
	if !a.emu.Loaded() {
		fmt.Println("No ROM. Use 'load <file>' first.")
		fmt.Println()
		return true
	}

	return false
}
