package main

import (
	"fmt"
	"io"
)

type cmdFunc func(app *App, args []string) error

var cmds = map[string]cmdFunc{
	"help": func(_ *App, _ []string) error {
		cmdHelp()
		return nil
	},

	"load": func(app *App, args []string) error {
		app.cmdLoad(args)
		return nil
	},

	"step": func(app *App, args []string) error {
		app.cmdStep(args)
		return nil
	},

	"regs": func(app *App, _args []string) error {
		app.cmdRegs()
		return nil
	},

	"peek": func(app *App, args []string) error {
		app.cmdPeek(args)
		return nil
	},

	"draw": func(app *App, _args []string) error {
		app.cmdDraw()
		return nil
	},

	"dis": func(app *App, _args []string) error {
		app.cmdDis()
		return nil
	},

	"info": func(app *App, _args []string) error {
		app.cmdInfo()
		return nil
	},

	"exit": func(_ *App, _ []string) error { return io.EOF },
	"quit": func(_ *App, _ []string) error { return io.EOF },
}

func cmdHelp() {
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
