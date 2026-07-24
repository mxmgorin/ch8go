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

	"break": func(app *App, args []string) error {
		app.cmdBreak(args)
		return nil
	},

	"breaks": func(app *App, _args []string) error {
		app.cmdBreaks()
		return nil
	},

	"delete": func(app *App, args []string) error {
		app.cmdDelete(args)
		return nil
	},

	"continue": func(app *App, args []string) error {
		app.cmdContinue(args)
		return nil
	},

	"c": func(app *App, args []string) error {
		app.cmdContinue(args)
		return nil
	},

	"mem": func(app *App, args []string) error {
		app.cmdMem(args)
		return nil
	},

	"keydown": func(app *App, args []string) error {
		app.cmdKey(args, true)
		return nil
	},

	"keyup": func(app *App, args []string) error {
		app.cmdKey(args, false)
		return nil
	},

	"keys": func(app *App, _args []string) error {
		app.cmdKeys()
		return nil
	},

	"watch": func(app *App, args []string) error {
		app.cmdWatch(args)
		return nil
	},

	"watches": func(app *App, _args []string) error {
		app.cmdWatches()
		return nil
	},

	"unwatch": func(app *App, args []string) error {
		app.cmdUnwatch(args)
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
  mem <addr> [n]  Hex-dump n bytes of memory from addr (default 64)
  break <addr>    Set a breakpoint at addr (hex 0x300 or decimal)
  breaks          List breakpoints
  delete [addr]   Remove a breakpoint (or all if omitted)
  watch <t>       Watch a memory addr or register (v0-vF); break on change
  watches         List watchpoints
  unwatch [t]     Remove a watchpoint (or all if omitted)
  continue [n]    Run until a breakpoint/watchpoint (or n steps max); alias: c
  keydown <hex>   Press a key (0-F)
  keyup <hex>     Release a key (0-F)
  keys            List currently pressed keys
  quit            Exit`)
	fmt.Println()
}
