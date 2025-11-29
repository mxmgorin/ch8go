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
	Pixels: [4]Color{
		{0, 0, 0},
		{255, 255, 255},
		{64, 64, 64},    // dark gray
		{192, 192, 192}, // light gray
	},
	Buzzer:  Color{255, 255, 255},
	Silence: Color{0, 0, 0},
}

type Color [3]byte

func (c Color) ToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
}

type Palette struct {
	Pixels  [4]Color
	Buzzer  Color
	Silence Color
}

type FrameBuffer struct {
	Pixels     []byte
	SoundColor Color
	Width      int
	Height     int
	BPP        int
}

func newFrameBuffer(w, h, bpp int) FrameBuffer {
	return FrameBuffer{Pixels: make([]byte, w*h*bpp), Width: w, Height: h, BPP: bpp}
}

func (fb *FrameBuffer) Pitch() int {
	return fb.Width * fb.BPP
}

type Painter interface {
	Init(w, h int) error
	Paint(fb *FrameBuffer)
	Destroy()
}

type App struct {
	VM          *chip8.VM
	DB          *db.DB
	ROMHash     string
	Palette     Palette
	frameBuffer FrameBuffer
	painter     Painter
	lastTime    time.Time
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

	return &App{
		DB:          db,
		VM:          vm,
		frameBuffer: newFrameBuffer(w, h, 4),
		painter:     painter,
		lastTime:    time.Now(),
		Palette:     DefaultPalette,
	}, nil
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
	a.Palette = DefaultPalette
	a.ROMHash = db.SHA1Of(rom)
	if err := a.VM.LoadROM(rom); err != nil {
		return 0, err
	}

	romInfo := a.ROMInfo()
	tickrate := 0

	for i := range romInfo.Platforms {
		id := romInfo.Platforms[i]
		if id != "megachip8" { // not supported
			platform := a.DB.FindPlatform(id)

			if platform != nil {
				quirks := chip8.Quirks{
					Shift:       platform.Quirks.Shift,
					MemIncIByX:  platform.Quirks.MemoryIncrementByX,
					MemLeaveI:   platform.Quirks.MemoryLeaveIUnchanged,
					Wrap:        platform.Quirks.Wrap,
					Jump:        platform.Quirks.Jump,
					VBlankWait:  platform.Quirks.VBlank,
					VFReset:     platform.Quirks.Logic,
					ScaleScroll: platform.Quirks.ScaleScroll,
				}
				a.VM.SetQuirks(quirks)
				fmt.Println("Quirks:", platform.ID)

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

	colors := romInfo.Colors
	if colors != nil && colors.Pixels != nil {
		if err := a.SetPalette(colors.Pixels, colors.Buzzer, colors.Silence); err != nil {
			fmt.Println("Failed to set palette")
		}
	}

	if tickrate > 0 {
		a.VM.SetTickrate(tickrate)
		fmt.Println("Tickrate:", tickrate)
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
	for i := range a.VM.Display.Height * a.VM.Display.Width {
		p0 := a.VM.Display.Planes[0][i]
		p1 := a.VM.Display.Planes[1][i]
		colorIdx := (p1 << 1) | p0
		color := a.Palette.Pixels[colorIdx]

		idx := i * a.frameBuffer.BPP
		a.frameBuffer.Pixels[idx] = color[0]
		a.frameBuffer.Pixels[idx+1] = color[1]
		a.frameBuffer.Pixels[idx+2] = color[2]
		a.frameBuffer.Pixels[idx+3] = 255
	}

	if a.VM.Buzzer() {
		a.frameBuffer.SoundColor = a.Palette.Buzzer
	} else {
		a.frameBuffer.SoundColor = a.Palette.Silence
	}

	a.painter.Paint(&a.frameBuffer)
}

func (a *App) SetColor(index int, hex string) error {
	color, err := ParseHexColor(hex)

	if err != nil {
		return err
	}

	a.Palette.Pixels[index] = color

	return nil
}

func (a *App) SetPalette(colors []string, buzzer, silence string) error {
	for i := 0; i < len(a.Palette.Pixels) && i < len(colors); i++ {
		a.SetColor(i, colors[i])
	}

	if buzzer != "" {
		color, err := ParseHexColor(buzzer)

		if err != nil {
			return nil
		}

		a.Palette.Buzzer = color
	}

	if silence != "" {
		color, err := ParseHexColor(silence)

		if err != nil {
			return nil
		}

		a.Palette.Buzzer = color
	}

	fmt.Println("Palette:", colors, buzzer, silence)

	return nil
}
func ParseHexColor(s string) (Color, error) {
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
