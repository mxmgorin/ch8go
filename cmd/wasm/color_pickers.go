package main


import (
	"syscall/js"

	"github.com/mxmgorin/ch8go/pkg/host"
)

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
