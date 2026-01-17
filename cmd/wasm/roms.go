//go:build js && wasm

package main

import (
	"syscall/js"
)

type Rom struct {
	Path string
	Name string
}

var Roms = []Rom{
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

func populateROMs(this js.Value, args []js.Value) any {
	selectEl := js.Global().Get("document").Call("getElementById", "roms")
	selectEl.Set("innerHTML", "")

	for _, r := range Roms {
		opt := js.Global().Get("document").Call("createElement", "option")
		opt.Set("value", r.Path)
		opt.Set("textContent", r.Name)
		selectEl.Call("appendChild", opt)
	}

	return nil
}

// Set ROM info in overlay
func setROMInfo(text string) {
	doc := js.Global().Get("document")
	info := doc.Call("getElementById", "info-overlay")
	info.Set("innerHTML", text)
}
