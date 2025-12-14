//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strconv"
	"syscall/js"
	"unsafe"

	"github.com/mxmgorin/ch8go/pkg/chip8"
	"github.com/mxmgorin/ch8go/pkg/host"
)

var (
	audioBuf = make([]float32, 0)
	app      App
	keymap   = map[string]chip8.Key{
		"1": chip8.Key1, "2": chip8.Key2, "3": chip8.Key3, "4": chip8.KeyC,
		"q": chip8.Key4, "w": chip8.Key5, "e": chip8.Key6, "r": chip8.KeyD,
		"a": chip8.Key7, "s": chip8.Key8, "d": chip8.Key9, "f": chip8.KeyE,
		"z": chip8.KeyA, "x": chip8.Key0, "c": chip8.KeyB, "v": chip8.KeyF,
		"ArrowUp": chip8.Key5, "ArrowDown": chip8.Key8, "ArrowLeft": chip8.Key7, "ArrowRight": chip8.Key9,
		" ": chip8.Key6,
	}
	roms = []struct {
		Path string
		Name string
	}{
		{"roms/xo/skyward.ch8", "Skyward"},
		{"roms/xo/superneatboy.ch8", "Super Neat Boy"},
		{"roms/xo/garlicscape.ch8", "Garlic Scape"},
		{"roms/xo/octoma.ch8", "Octoma"},
		{"roms/xo/t8nks.ch8", "T8nks"},
		{"roms/xo/octopeg.ch8", "Octopeg"},
		{"roms/ch/danm8ku.ch8", "Danm8ku"},
		{"roms/ch/octogon.ch8", "Octogon"},
		{"roms/ch/supersquare.ch8", "Super Square"},
		{"roms/ch/down8.ch8", "Down8"},
		{"roms/ch/slipperyslope.ch8", "Slippery Slope"},
		{"roms/ch/rockto.ch8", "Rockto"},
		{"roms/ch/sub8.ch8", "Sub8"},
		{"roms/ch/DVN8.ch8", "DVN8"},
		{"roms/ch/flightrunner.ch8", "Flight Runner"},
		{"roms/ch/glitchGhost.ch8", "Glitch Ghost"},
		{"roms/ch/turnover77.ch8", "Turn Over 77"},
		{"roms/ch/blackrainbow.ch8", "Black Rainbow"},
		{"roms/ch/binding.ch8", "Binding"},
		{"roms/ch/br8kout.ch8", "Br8kout"},
		{"roms/ch/spacejam.ch8", "Space Jam"},
		{"roms/ch/octovore.ch8", "Octovore"},
		{"roms/ch/INVADERS", "Invaders"},
		{"roms/ch/TETRIS", "Tetris"},
		{"roms/ch/snake.ch8", "Snake"},
		{"roms/ch/TANK", "Tank"},
		{"roms/xo/D8GN.ch8", "D8GN"},
		{"roms/xo/civiliz8n.ch8", "Civiliz8n"},
		{"roms/xo/clostro.ch8", "Clostro"},
		{"roms/xo/sneaksurround.ch8", "Sneak Surround"},
		{"roms/xo/chickenScratch.ch8", "Chicken Scratch"},
		{"roms/sc/ANT", "Ant"},
		{"roms/sc/sweetcopter.ch8", "Sw8Copter"},
		{"roms/xo/tapeworm.ch8", "Tapeworm"},
		{"roms/xo/snake.ch8", "xSnake"},
		{"roms/xo/alien-inv8sion.ch8", "Alien Inv8sion (Timendus)"},
	}
)

type Painter struct {
	ctx           js.Value
	imageData     js.Value
	screen        js.Value
	canvas        js.Value
	screenBgColor *host.Color
	width         int
	height        int
}

func (p *Painter) setScreenBg(color host.Color) {
	if p.screenBgColor == nil || *p.screenBgColor != color {
		hex := color.ToHex()
		p.screen.Get("style").Set("background", hex)
		p.screenBgColor = &color
	}
}

func newPainter(w, h int) (Painter, error) {
	p := Painter{}
	p.width = w
	p.height = h
	doc := js.Global().Get("document")

	p.canvas = doc.Call("getElementById", "chip8-canvas")
	p.canvas.Set("width", w)
	p.canvas.Set("height", h)

	p.screen = doc.Call("getElementById", "chip8-screen")

	scaleInput := doc.Call("getElementById", "scaleInput")
	value := scaleInput.Get("value").String()
	p.setScale(value)
	scaleInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := scaleInput.Get("value").String()
		p.setScale(value)
		return nil
	}))

	p.ctx = p.canvas.Call("getContext", "2d")
	p.imageData = p.ctx.Call("createImageData", w, h)
	return p, nil
}

func (p *Painter) setScale(value string) {
	scale, err := strconv.Atoi(value)
	if err != nil {
		slog.Error("Invalid scale:", "scale", value)
		return
	}

	canvasStyle := p.canvas.Get("style")
	canvasStyle.Set("transform", fmt.Sprintf("scale(%d)", scale))

	screenStyle := p.screen.Get("style")
	screenStyle.Set("width", fmt.Sprintf("%dpx", p.width*scale))
	screenStyle.Set("height", fmt.Sprintf("%dpx", p.height*scale))
}

func (p *Painter) Paint(fb *host.FrameBuffer) {
	p.setScreenBg(fb.SoundColor)
	js.CopyBytesToJS(p.imageData.Get("data"), fb.Pixels)
	p.ctx.Call("putImageData", p.imageData, 0, 0)
}

type ColorPickers [4]js.Value

func newColorPickers(doc js.Value, pal *host.Palette) ColorPickers {
	pickers := ColorPickers{}
	pickers[0] = doc.Call("getElementById", "bgPicker")
	pickers[1] = doc.Call("getElementById", "fgPicker")
	pickers[2] = doc.Call("getElementById", "c3Picker")
	pickers[3] = doc.Call("getElementById", "c4Picker")

	for i := range pickers {
		id := i
		pickers[id].Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
			color := pickers[id].Get("value").String()
			pal.SetColor(id, color)
			return nil
		}))
	}

	return pickers
}

func (cp *ColorPickers) setColors(colors *[16]host.Color) {
	for i := range cp {
		cp[i].Set("value", colors[i].ToHex())
	}
}

type KeyEvent struct {
	Key     chip8.Key
	Pressed bool
}

type ConfForm struct {
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

func newConf(doc js.Value, vm *chip8.VM) ConfForm {
	conf := ConfForm{}

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

func (c *ConfForm) setTickrate(tr int) {
	c.tickrateInput.Set("value", tr)
}

func (c *ConfForm) setQuirks(quirks chip8.Quirks) {
	c.shiftInput.Set("checked", js.ValueOf(quirks.Shift))
	c.wrapInput.Set("checked", js.ValueOf(quirks.Wrap))
	c.incIByXInput.Set("checked", js.ValueOf(quirks.MemIncIByX))
	c.leaveIInput.Set("checked", js.ValueOf(quirks.MemLeaveI))
	c.jumpInput.Set("checked", js.ValueOf(quirks.Jump))
	c.waitVBlankInput.Set("checked", js.ValueOf(quirks.WaitVBlank))
	c.resetFlagInput.Set("checked", js.ValueOf(quirks.ResetFlag))
	c.scaleScrollInput.Set("checked", js.ValueOf(quirks.ScaleScroll))
}

type App struct {
	runFrameFunc    js.Func
	emu             *host.Emu
	colorPickers    ColorPickers
	painter         Painter
	confForm        ConfForm
	togglePauseIcon js.Value
	pauseOverlay    js.Value
	keyChan         chan KeyEvent
}

func newApp() App {
	emu, err := host.NewEmu()
	if err != nil {
		log.Fatal(err)
	}

	size := emu.VM.Display.Size()
	painter, err := newPainter(size.Width, size.Height)
	if err != nil {
		log.Fatal(err)
	}

	// Export ROM loader
	js.Global().Set("chip8_loadROM", js.FuncOf(loadROM))

	// Handle keyboard
	win := js.Global().Get("window")
	win.Call("addEventListener", "keydown", js.FuncOf(onKeyDown))
	win.Call("addEventListener", "keyup", js.FuncOf(onKeyUp))

	// Palette
	doc := js.Global().Get("document")
	colorPickers := newColorPickers(doc, &emu.Palette)

	// Conf
	conf := newConf(doc, emu.VM)

	// Audio
	js.Global().Set("fillAudio", js.FuncOf(fillAudio))
	js.Global().Set("startAudio", js.FuncOf(startAudio))

	// Pause, resume
	pauseOverlay := js.Global().Get("document").Call("getElementById", "pause-overlay")
	togglePauseIcon := doc.Call("getElementById", "toggle-pause-icon")
	togglePauseBtn := doc.Call("getElementById", "toggle-pause-btn")
	togglePauseBtn.Call("addEventListener", "click", js.FuncOf(togglePause))

	// Roms
	js.Global().Set("fillROMs", js.FuncOf(fillROMs))

	// Animation loop (must persist function or GC will kill it)
	runFrameFunc := js.FuncOf(runFrame)

	return App{
		emu:             emu,
		runFrameFunc:    runFrameFunc,
		colorPickers:    colorPickers,
		painter:         painter,
		confForm:        conf,
		togglePauseIcon: togglePauseIcon,
		pauseOverlay:    pauseOverlay,
		keyChan:         make(chan KeyEvent, 32),
	}
}

// Run main loop
func (a *App) run() {
	js.Global().Call("requestAnimationFrame", a.runFrameFunc)
	// Keep WASM alive
	select {}
}

// Set ROM info in overlay
func setROMInfo() {
	text := app.emu.ROMInfo()
	doc := js.Global().Get("document")
	info := doc.Call("getElementById", "info-overlay")
	info.Set("innerHTML", text)
}

func loadROM(this js.Value, args []js.Value) any {
	jsBuff := args[0]
	name := args[1].String()
	buf := make([]byte, jsBuff.Length())
	js.CopyBytesToGo(buf, jsBuff)
	_, err := app.emu.LoadROM(buf, filepath.Ext(name))
	if err != nil {
		fmt.Println(err)
	}

	app.colorPickers.setColors(&app.emu.Palette.Pixels)
	app.confForm.setTickrate(app.emu.VM.Tickrate())
	app.confForm.setQuirks(app.emu.VM.CPU.Quirks)
	setROMInfo()

	return nil
}

func onKeyDown(this js.Value, args []js.Value) any {
	return onKey(args[0], true)
}

func onKeyUp(this js.Value, args []js.Value) any {
	return onKey(args[0], false)
}

func onKey(event js.Value, pressed bool) any {
	// Detect if user is typing into an input, textarea, or contenteditable
	target := event.Get("target")
	nodeName := target.Get("nodeName").String()
	isContentEditable := target.Get("isContentEditable").Truthy()

	if nodeName == "INPUT" || nodeName == "TEXTAREA" || isContentEditable {
		// Don't block keypresses on input fields
		return nil
	}

	key := event.Get("key").String()
	if k, ok := keymap[key]; ok {
		app.keyChan <- KeyEvent{Key: k, Pressed: pressed}
		event.Call("preventDefault")
	}
	return nil
}

func startAudio(this js.Value, args []js.Value) any {
	size := args[0].Int()
	audioBuf = make([]float32, size)
	return nil
}

func fillAudio(this js.Value, args []js.Value) any {
	out := args[0] // JS Float32Array
	freq := args[1].Float()
	app.emu.VM.Audio.Output(audioBuf, freq)

	outBuffer := js.Global().Get("Uint8Array").New(
		out.Get("buffer"),
		out.Get("byteOffset"),
		out.Get("byteLength"),
	)
	bufPointer := unsafe.Pointer(&audioBuf[0])
	byteBuf := unsafe.Slice((*byte)(bufPointer), len(audioBuf)*4)
	js.CopyBytesToJS(outBuffer, byteBuf)

	return nil
}

func runFrame(this js.Value, args []js.Value) any {
	drainChan(app.keyChan, handleKey)
	fb := app.emu.RunFrame()
	app.painter.Paint(fb)
	// Schedule next frame
	js.Global().Call("requestAnimationFrame", app.runFrameFunc)
	return nil
}

func handleKey(evt KeyEvent) {
	if evt.Pressed {
		app.emu.VM.Keypad.Press(evt.Key)
	} else {
		app.emu.VM.Keypad.Release(evt.Key)
	}
}

func fillROMs(this js.Value, args []js.Value) any {
	selectEl := js.Global().Get("document").Call("getElementById", "roms")
	selectEl.Set("innerHTML", "")

	for _, r := range roms {
		opt := js.Global().Get("document").Call("createElement", "option")
		opt.Set("value", r.Path)
		opt.Set("textContent", r.Name)
		selectEl.Call("appendChild", opt)
	}

	return nil
}

func togglePause(this js.Value, args []js.Value) any {
	app.emu.Paused = !app.emu.Paused

	if app.emu.Paused {
		app.pauseOverlay.Get("classList").Call("add", "active")
		app.togglePauseIcon.Set("src", "play-icon.svg")
	} else {
		app.pauseOverlay.Get("classList").Call("remove", "active")
		app.togglePauseIcon.Set("src", "pause-icon.svg")
	}

	return nil
}

func drainChan[T any](ch <-chan T, fn func(T)) {
	for {
		select {
		case v := <-ch:
			fn(v)
		default:
			return
		}
	}
}

func main() {
	slog.Info("ch8go WASM")
	app = newApp()
	app.run()
}
