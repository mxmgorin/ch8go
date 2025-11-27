package app

import (
	"fmt"
	"os"

	"github.com/mxmgorin/ch8go/app/db"
	"github.com/mxmgorin/ch8go/chip8"
)

type App struct {
	VM      *chip8.VM
	DB      *db.DB
	ROMHash string
	rgbaBuf []byte
	painter Painter
}

type Painter interface {
	Init(w, h int) error
	Paint(rgbaBuf []byte, w, h int)
	Destroy()
}

func NewApp(painter Painter) (*App, error) {
	db, err := db.NewDB()

	if err != nil {
		fmt.Println("Failed to create DB:", err)
	}

	vm := chip8.NewVM()
	w := vm.Display.Width
	h := vm.Display.Height

	if painter != nil {
		if err := painter.Init(w, h); err != nil {
			return nil, err
		}
	}

	return &App{DB: db, VM: vm, rgbaBuf: make([]byte, w*h*4), painter: painter}, nil
}

func (a *App) Quit() {
	a.painter.Destroy()
}

func (a *App) HasROM() bool {
	return a.ROMHash != ""
}

// Should be run at 60 fps
func (a *App) PaintFrame() {
	if a.VM.RunFrame() && a.HasROM() {
		a.Paint()
	}
}

func (a *App) ReadROM(path string) (int, error) {
	rom, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return a.LoadROM(rom)
}

func (a *App) LoadROM(rom []byte) (int, error) {
	a.ROMHash = db.SHA1Of(rom)
	if err := a.VM.LoadROM(rom); err != nil {
		return 0, err
	}

	romInfo := a.ROMInfo()
	tickrate := 0

	for i := range romInfo.Platforms {
		id := romInfo.Platforms[i]
		if id != "xochip" && id != "megachip8" {
			platform := a.DB.FindPlatform(id)

			if platform != nil {
				quirks := chip8.Quirks{
					Shift:      platform.Quirks.Shift,
					MemIncIByX: platform.Quirks.MemoryIncrementByX,
					MemLeaveI:  platform.Quirks.MemoryLeaveIUnchanged,
					Wrap:       platform.Quirks.Wrap,
					Jump:       platform.Quirks.Jump,
					VBlankWait: platform.Quirks.VBlank,
					VFReset:    platform.Quirks.Logic,
				}
				a.VM.SetQuirks(quirks)
				fmt.Println("Set quirks", platform.ID)

				if platform.DefaultTickrate > 0 {
					tickrate = platform.DefaultTickrate
				}
				break
			}
		}
	}

	if romInfo.Tickrate > 0 {
		tickrate = romInfo.Tickrate
	}

	if tickrate > 0 {
		a.VM.SetTickrate(tickrate)
		fmt.Println("Set tickrate", tickrate)
	}

	return len(rom), nil
}

func (a *App) ROMInfo() *db.RomDto {
	program := a.DB.FindProgram(a.ROMHash)
	if program == nil {
		return nil // Unknown ROM
	}
	rom := program.ROMs[a.ROMHash]

	return &rom
}

func (a *App) Paint() {
	pixels := a.VM.Display.Pixels

	for i := range pixels {
		v := byte(0)
		if pixels[i] != 0 {
			v = 255
		}
		idx := i * 4
		a.rgbaBuf[idx] = v
		a.rgbaBuf[idx+1] = v
		a.rgbaBuf[idx+2] = v
		a.rgbaBuf[idx+3] = 255
	}

	a.painter.Paint(a.rgbaBuf, a.VM.Display.Width, a.VM.Display.Height)
}
