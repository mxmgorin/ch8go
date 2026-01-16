package main

import (
	"log/slog"
	"syscall/js"
	"strconv"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

type ConfOverlay struct {
	tickrateInput    js.Value
	shiftInput       js.Value
	incIByXInput     js.Value
	leaveIInput      js.Value
	wrapInput        js.Value
	jumpInput        js.Value
	waitVBlankInput  js.Value
	resetFlagInput   js.Value
	scaleScrollInput js.Value
}

func newConf(doc js.Value, vm *chip8.VM) ConfOverlay {
	conf := ConfOverlay{}

	conf.tickrateInput = doc.Call("getElementById", "tickrateInput")
	conf.tickrateInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.tickrateInput.Get("value").String()
		tickrate, err := strconv.Atoi(value)
		if err != nil {
			slog.Error("Invalid tickrate:", "tickrate", value)
			return nil
		}
		vm.SetTickrate(tickrate)

		return nil
	}))

	conf.shiftInput = doc.Call("getElementById", "shiftInput")
	conf.shiftInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.shiftInput.Get("checked").Bool()
		vm.CPU.Quirks.Shift = value

		return nil
	}))

	conf.wrapInput = doc.Call("getElementById", "wrapInput")
	conf.wrapInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.wrapInput.Get("checked").Bool()
		vm.CPU.Quirks.Wrap = value

		return nil
	}))

	conf.incIByXInput = doc.Call("getElementById", "incIbyXInput")
	conf.incIByXInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.incIByXInput.Get("checked").Bool()
		vm.CPU.Quirks.MemIncIByX = value

		return nil
	}))

	conf.leaveIInput = doc.Call("getElementById", "leaveIInput")
	conf.leaveIInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.leaveIInput.Get("checked").Bool()
		vm.CPU.Quirks.MemLeaveI = value

		return nil
	}))

	conf.jumpInput = doc.Call("getElementById", "jumpInput")
	conf.jumpInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.jumpInput.Get("checked").Bool()
		vm.CPU.Quirks.Jump = value

		return nil
	}))

	conf.waitVBlankInput = doc.Call("getElementById", "vblankInput")
	conf.waitVBlankInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.waitVBlankInput.Get("checked").Bool()
		vm.CPU.Quirks.WaitVBlank = value

		return nil
	}))

	conf.resetFlagInput = doc.Call("getElementById", "resetFInput")
	conf.resetFlagInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.resetFlagInput.Get("checked").Bool()
		vm.CPU.Quirks.ResetFlag = value

		return nil
	}))

	conf.scaleScrollInput = doc.Call("getElementById", "scaleScrollInput")
	conf.scaleScrollInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := conf.scaleScrollInput.Get("checked").Bool()
		vm.CPU.Quirks.ScaleScroll = value

		return nil
	}))

	return conf
}

func (c *ConfOverlay) setTickrate(tr int) {
	c.tickrateInput.Set("value", tr)
}

func (c *ConfOverlay) setQuirks(quirks chip8.Quirks) {
	c.shiftInput.Set("checked", js.ValueOf(quirks.Shift))
	c.wrapInput.Set("checked", js.ValueOf(quirks.Wrap))
	c.incIByXInput.Set("checked", js.ValueOf(quirks.MemIncIByX))
	c.leaveIInput.Set("checked", js.ValueOf(quirks.MemLeaveI))
	c.jumpInput.Set("checked", js.ValueOf(quirks.Jump))
	c.waitVBlankInput.Set("checked", js.ValueOf(quirks.WaitVBlank))
	c.resetFlagInput.Set("checked", js.ValueOf(quirks.ResetFlag))
	c.scaleScrollInput.Set("checked", js.ValueOf(quirks.ScaleScroll))
}
