package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/mxmgorin/ch8go/pkg/chip8"
	"github.com/mxmgorin/ch8go/pkg/host"
)

const maxRunSteps = 5_000_000

type App struct {
	emu     *host.Emu
	painter ASCIIPainter
	breaks  map[uint16]bool
	watches map[chip8.Watch]bool
}

func newApp() App {
	a, _ := host.NewEmu()
	return App{
		emu:     a,
		painter: ASCIIPainter{},
		breaks:  map[uint16]bool{},
		watches: map[chip8.Watch]bool{},
	}
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
		a.emu.VM.Poll() // clear pending VBlank so WaitVBlank ROMs advance
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

// parseAddr accepts hex ("0x300") or decimal ("768").
func parseAddr(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 0, 16)
	return uint16(v), err
}

func (a *App) cmdBreak(args []string) {
	if a.loaded() {
		return
	}

	if len(args) < 2 {
		fmt.Println("Usage: break <addr>   e.g. break 0x300")
		fmt.Println()
		return
	}

	addr, err := parseAddr(args[1])
	if err != nil {
		fmt.Println("Invalid address:", args[1])
		return
	}

	a.breaks[addr] = true
	fmt.Printf("Breakpoint set at %04X.\n\n", addr)
}

func (a *App) cmdBreaks() {
	if len(a.breaks) == 0 {
		fmt.Println("No breakpoints.")
		fmt.Println()
		return
	}

	addrs := make([]uint16, 0, len(a.breaks))
	for addr := range a.breaks {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

	for _, addr := range addrs {
		fmt.Printf("  %04X\n", addr)
	}
	fmt.Println()
}

func (a *App) cmdDelete(args []string) {
	if len(args) < 2 {
		a.breaks = map[uint16]bool{}
		fmt.Println("All breakpoints cleared.")
		fmt.Println()
		return
	}

	addr, err := parseAddr(args[1])
	if err != nil {
		fmt.Println("Invalid address:", args[1])
		return
	}

	delete(a.breaks, addr)
	fmt.Printf("Breakpoint at %04X removed.\n\n", addr)
}

func (a *App) cmdContinue(args []string) {
	if a.loaded() {
		return
	}

	max := maxRunSteps
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil {
			max = n
		} else {
			fmt.Println("Invalid number:", args[1])
			return
		}
	}

	watches := make([]chip8.Watch, 0, len(a.watches))
	for w := range a.watches {
		watches = append(watches, w)
	}

	res := a.emu.VM.Run(a.breaks, watches, max)
	switch res.Reason {
	case chip8.StopBreakpoint:
		fmt.Printf("Hit breakpoint after %d steps.\n", res.Steps)
		fmt.Println(a.emu.VM.PeekNext())
	case chip8.StopWatch:
		fmt.Printf("Watchpoint %s changed %02X -> %02X after %d steps.\n",
			res.WatchDesc, res.WatchOld, res.WatchNew, res.Steps)
		fmt.Println(a.emu.VM.PeekNext())
	default:
		fmt.Printf("Ran %d steps, no stop (cap %d).\n", res.Steps, max)
	}

	fmt.Println()
}

func (a *App) cmdMem(args []string) {
	if a.loaded() {
		return
	}

	if len(args) < 2 {
		fmt.Println("Usage: mem <addr> [n]   e.g. mem 0x200 64")
		fmt.Println()
		return
	}

	addr, err := parseAddr(args[1])
	if err != nil {
		fmt.Println("Invalid address:", args[1])
		return
	}

	n := 64
	if len(args) >= 3 {
		if v, err := strconv.Atoi(args[2]); err == nil && v > 0 {
			n = v
		} else {
			fmt.Println("Invalid count:", args[2])
			return
		}
	}

	end := int(addr) + n
	if end > chip8.MemorySize {
		end = chip8.MemorySize
	}

	for base := int(addr); base < end; base += 16 {
		var hex, ascii strings.Builder
		for off := 0; off < 16; off++ {
			cur := base + off
			if cur >= end {
				hex.WriteString("   ")
				continue
			}
			b := a.emu.VM.Memory.Read(uint16(cur))
			fmt.Fprintf(&hex, "%02X ", b)
			if b >= 0x20 && b < 0x7f {
				ascii.WriteByte(b)
			} else {
				ascii.WriteByte('.')
			}
		}
		fmt.Printf("%04X  %s|%s|\n", base, hex.String(), ascii.String())
	}
	fmt.Println()
}

func parseKey(s string) (chip8.Key, error) {
	v, err := strconv.ParseUint(s, 16, 8)
	if err != nil || v > 0xF {
		return 0, fmt.Errorf("invalid key %q (use 0-F)", s)
	}
	return chip8.Key(v), nil
}

func (a *App) cmdKey(args []string, down bool) {
	if a.loaded() {
		return
	}

	verb := "keyup"
	if down {
		verb = "keydown"
	}
	if len(args) < 2 {
		fmt.Printf("Usage: %s <hex>   e.g. %s 5\n\n", verb, verb)
		return
	}

	k, err := parseKey(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	if down {
		a.emu.VM.Keypad.Press(k)
	} else {
		a.emu.VM.Keypad.Release(k)
	}
	fmt.Printf("Key %X %s.\n\n", byte(k), verb[3:])
}

func (a *App) cmdKeys() {
	if a.loaded() {
		return
	}

	pressed := []string{}
	for k := chip8.Key(0); k < chip8.KeyCount; k++ {
		if a.emu.VM.Keypad.IsPressed(k) {
			pressed = append(pressed, fmt.Sprintf("%X", byte(k)))
		}
	}

	if len(pressed) == 0 {
		fmt.Println("No keys pressed.")
	} else {
		fmt.Printf("Pressed: %s\n", strings.Join(pressed, " "))
	}
	fmt.Println()
}

func parseWatch(s string) (chip8.Watch, error) {
	if len(s) >= 2 && (s[0] == 'v' || s[0] == 'V') {
		n, err := strconv.ParseUint(s[1:], 16, 8)
		if err != nil || n > 15 {
			return chip8.Watch{}, fmt.Errorf("invalid register %q (use v0-vF)", s)
		}
		return chip8.Watch{Reg: true, Addr: uint16(n)}, nil
	}

	addr, err := parseAddr(s)
	if err != nil {
		return chip8.Watch{}, fmt.Errorf("invalid target %q (use <addr> or v<x>)", s)
	}
	return chip8.Watch{Reg: false, Addr: addr}, nil
}

func (a *App) cmdWatch(args []string) {
	if a.loaded() {
		return
	}

	if len(args) < 2 {
		fmt.Println("Usage: watch <addr>|v<x>   e.g. watch 0x300  |  watch v5")
		fmt.Println()
		return
	}

	w, err := parseWatch(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	a.watches[w] = true
	fmt.Printf("Watching %s.\n\n", w)
}

func (a *App) cmdWatches() {
	if len(a.watches) == 0 {
		fmt.Println("No watchpoints.")
		fmt.Println()
		return
	}

	list := make([]chip8.Watch, 0, len(a.watches))
	for w := range a.watches {
		list = append(list, w)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Reg != list[j].Reg {
			return !list[i].Reg // memory before registers
		}
		return list[i].Addr < list[j].Addr
	})

	for _, w := range list {
		fmt.Printf("  %s\n", w)
	}
	fmt.Println()
}

func (a *App) cmdUnwatch(args []string) {
	if len(args) < 2 {
		a.watches = map[chip8.Watch]bool{}
		fmt.Println("All watchpoints cleared.")
		fmt.Println()
		return
	}

	w, err := parseWatch(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	delete(a.watches, w)
	fmt.Printf("Stopped watching %s.\n\n", w)
}

func (a *App) loaded() bool {
	if !a.emu.Loaded() {
		fmt.Println("No ROM. Use 'load <file>' first.")
		fmt.Println()
		return true
	}

	return false
}
