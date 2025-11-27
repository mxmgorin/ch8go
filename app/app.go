package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mxmgorin/ch8go/app/db"
	"github.com/mxmgorin/ch8go/chip8"
)

var DefaultPalette = Palette{
	Color{0, 0, 0},       // background
	Color{255, 255, 255}, // foreground
}

type Color [3]byte
type Palette [2]Color // [foreground/background][RGB]

type App struct {
	VM       *chip8.VM
	DB       *db.DB
	ROMHash  string
	rgbaBuf  []byte
	painter  Painter
	lastTime time.Time
	palette  Palette
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

	return &App{DB: db, VM: vm, rgbaBuf: make([]byte, w*h*4), painter: painter, lastTime: time.Now(), palette: DefaultPalette}, nil
}

func (a *App) Quit() {
	a.painter.Destroy()
}

func (a *App) HasROM() bool {
	return a.ROMHash != ""
}

func (a *App) PaintFrame() {
	now := time.Now()
	dt := now.Sub(a.lastTime).Seconds()
	a.lastTime = now

	if a.VM.RunFrame(dt) && a.HasROM() {
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
	a.palette = DefaultPalette
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

	if romInfo.Colors != nil && romInfo.Colors.Pixels != nil {
		a.SetPalette(romInfo.Colors.Pixels)
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
		color := a.palette[pixels[i]] // pixels[i] is 0 or 1
		idx := i * 4

		a.rgbaBuf[idx] = color[0]
		a.rgbaBuf[idx+1] = color[1]
		a.rgbaBuf[idx+2] = color[2]
		a.rgbaBuf[idx+3] = 255
	}

	a.painter.Paint(a.rgbaBuf, a.VM.Display.Width, a.VM.Display.Height)
}

func (a *App) SetPalette(colors []string) error {
	for i := 0; i < 2 && i < len(colors); i++ {
		color, err := parseHexColor(colors[i])

		if err != nil {
			return err
		}

		a.palette[i] = color
	}

	return nil
}

func parseHexColor(s string) (Color, error) {
	s = strings.TrimPrefix(s, "#")

	if len(s) != 6 {
		return Color{}, fmt.Errorf("invalid hex color: %q", s)
	}

	ri, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return Color{}, err
	}

	gi, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return Color{}, err
	}

	bi, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return Color{}, err
	}

	return Color{byte(ri), byte(gi), byte(bi)}, nil
}
