package app

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mxmgorin/ch8go/app/db"
	"github.com/mxmgorin/ch8go/chip8"
)

var DefaultPalette = Palette{
	Pixels: [16]Color{
		{0, 0, 0},       // #000000
		{255, 255, 255}, // #FFFFFF
		{170, 170, 170}, // #AAAAAA
		{85, 85, 85},    // #555555
		{255, 0, 0},     // #FF0000
		{0, 255, 0},     // #00FF00
		{0, 0, 255},     // #0000FF
		{255, 255, 0},   // #FFFF00
		{136, 0, 0},     // #880000
		{0, 136, 0},     // #008800
		{0, 0, 136},     // #000088
		{136, 136, 0},   // #888800
		{255, 0, 255},   // #FF00FF
		{0, 255, 255},   // #00FFFF
		{136, 0, 136},   // #880088
		{0, 136, 136},   // #008888
	},
	Buzzer:  Color{255, 255, 255},
	Silence: Color{0, 0, 0},
}

type Color [3]byte

func (c Color) ToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
}

type Palette struct {
	Pixels  [16]Color
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

func (fb *FrameBuffer) Hash() string {
	sum := sha256.Sum256(fb.Pixels)
	return fmt.Sprintf("%x", sum[:])
}

func (fb *FrameBuffer) SavePNG(path string) error {
	if fb.BPP != 4 {
		return fmt.Errorf("expected BPP=4 (RGBA), got %d", fb.BPP)
	}

	img := image.NewRGBA(image.Rect(0, 0, fb.Width, fb.Height))

	// fb.Pixels is already RGBA ordered → safe direct copy
	copy(img.Pix, fb.Pixels)

	os.MkdirAll(filepath.Dir(path), 0755)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

type App struct {
	VM          *chip8.VM
	DB          *db.DB
	ROMHash     string
	Palette     Palette
	Paused      bool
	frameBuffer FrameBuffer
	lastTime    time.Time
}

func NewApp() (*App, error) {
	db, err := db.NewDB()

	if err != nil {
		slog.Error("Failed to create DB", "err", err)
	}

	vm := chip8.NewVM()
	w := vm.Display.Width
	h := vm.Display.Height

	return &App{
		DB:          db,
		VM:          vm,
		frameBuffer: newFrameBuffer(w, h, 4),
		lastTime:    time.Now(),
		Palette:     DefaultPalette,
	}, nil
}

func (a *App) HasROM() bool {
	return a.ROMHash != ""
}

func (a *App) RunFrame() *FrameBuffer {
	now := time.Now()
	dt := now.Sub(a.lastTime).Seconds()
	a.lastTime = now

	return a.RunFrameDT(dt)
}

func (a *App) RunFrameDT(dt float64) *FrameBuffer {
	if a.HasROM() && !a.Paused {
		state := a.VM.RunFrame(dt)
		a.updateFrameBuffer(state)
	}
	return &a.frameBuffer
}

func (a *App) ReadROM(path string) (int, error) {
	rom, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return a.LoadROM(rom, filepath.Ext(path))
}

func (a *App) LoadROM(rom []byte, ext string) (int, error) {
	a.Palette = DefaultPalette
	a.ROMHash = db.SHA1Of(rom)
	len := len(rom)

	slog.Info("ROM loaded:", "size", len, "hash", a.ROMHash, "ext", ext)

	if err := a.VM.LoadROM(rom); err != nil {
		return 0, err
	}

	platform, ok := chip8.PlatformByExt[ext]
	if ok {
		conf, ok := chip8.ConfByPlatform[platform]
		if ok {
			a.VM.ApplyConf(conf)
		}
	}
	a.applyROMconf()

	return len, nil
}

func (a *App) ROMInfo() *db.RomDto {
	program := a.DB.FindProgram(a.ROMHash)
	if program == nil {
		return nil // Unknown ROM
	}
	rom := program.ROMs[a.ROMHash]

	return &rom
}

func (a *App) ROMDesc() string {
	program := a.DB.FindProgram(a.ROMHash)
	if program == nil {
		return "Unknown"
	}
	return program.Format()
}

func (a *App) applyROMconf() {
	romInfo := a.ROMInfo()
	if romInfo == nil {
		slog.Info("Unknown ROM")
		return
	}

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
				slog.Info("Quirks:", "platformID", platform.ID)

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
			slog.Error("Failed to set palette", "err", err)
		}
	}

	if tickrate > 0 {
		a.VM.SetTickrate(tickrate)
		slog.Info("Tickrate:", "val", tickrate)
	}
}

func (a *App) updateFrameBuffer(frameState chip8.FrameState) {
	if frameState.Dirty {
		for i := range a.VM.Display.Height * a.VM.Display.Width {
			colorIdx := 0
			for plane := range a.VM.Display.Planes {
				pixel := a.VM.Display.Planes[plane][i]
				colorIdx |= int(pixel) << plane
			}

			color := a.Palette.Pixels[colorIdx]
			idx := i * a.frameBuffer.BPP
			a.frameBuffer.Pixels[idx] = color[0]
			a.frameBuffer.Pixels[idx+1] = color[1]
			a.frameBuffer.Pixels[idx+2] = color[2]
			a.frameBuffer.Pixels[idx+3] = 255
		}
	}

	if frameState.Beep {
		a.frameBuffer.SoundColor = a.Palette.Buzzer
	} else {
		a.frameBuffer.SoundColor = a.Palette.Silence
	}
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

	slog.Info("Palette:", "colors", colors, "buzzer", buzzer, "silence", silence)

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
