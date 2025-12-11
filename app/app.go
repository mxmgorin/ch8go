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
		{0, 0, 0, 255},       // #000000
		{255, 255, 255, 255}, // #FFFFFF
		{170, 170, 170, 255}, // #AAAAAA
		{85, 85, 85, 255},    // #555555
		{255, 0, 0, 255},     // #FF0000
		{0, 255, 0, 255},     // #00FF00
		{0, 0, 255, 255},     // #0000FF
		{255, 255, 0, 255},   // #FFFF00
		{136, 0, 0, 255},     // #880000
		{0, 136, 0, 255},     // #008800
		{0, 0, 136, 255},     // #000088
		{136, 136, 0, 255},   // #888800
		{255, 0, 255, 255},   // #FF00FF
		{0, 255, 255, 255},   // #00FFFF
		{136, 0, 136, 255},   // #880088
		{0, 136, 136, 255},   // #008888
	},
	Buzzer:  Color{255, 255, 255, 255},
	Silence: Color{0, 0, 0, 255},
}

type Color [4]byte

func (c Color) ToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
}

type Palette struct {
	Pixels  [16]Color
	Buzzer  Color
	Silence Color
}

func (a *Palette) SetColor(index int, hex string) error {
	color, err := ParseHexColor(hex)

	if err != nil {
		return err
	}

	a.Pixels[index] = color

	return nil
}

func NewPalette(colors []string, buzzer, silence string) (p Palette, e error) {
	for i := 0; i < len(p.Pixels) && i < len(colors); i++ {
		p.SetColor(i, colors[i])
	}

	if buzzer != "" {
		color, err := ParseHexColor(buzzer)

		if err != nil {
			return p, err
		}

		p.Buzzer = color
	}

	if silence != "" {
		color, err := ParseHexColor(silence)

		if err != nil {
			return p, err
		}

		p.Buzzer = color
	}

	return p, nil
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

func (a *App) updateFrameBuffer(frameState chip8.FrameState) {
	if frameState.Dirty {
		size := a.VM.Display.Height * a.VM.Display.Width
		planes := a.VM.Display.Planes
		bpp := a.frameBuffer.BPP
		pixels := a.frameBuffer.Pixels
		palette := a.Palette.Pixels

		for i := range size {
			colorIdx := int(planes[0][i]) | int(planes[1][i])<<1 | int(planes[2][i])<<2 | int(planes[3][i])<<3
			idx := i * bpp
			copy(pixels[idx:idx+4], palette[colorIdx][:])
		}
	}

	if frameState.Beep {
		a.frameBuffer.SoundColor = a.Palette.Buzzer
	} else {
		a.frameBuffer.SoundColor = a.Palette.Silence
	}
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

	meta := a.ROMMeta()
	conf := a.ROMConf(meta, ext)
	a.VM.SetConf(conf)

	if meta != nil {
		colors := meta.Colors
		if colors != nil && colors.Pixels != nil {
			p, err := NewPalette(colors.Pixels, colors.Buzzer, colors.Silence)
			if err != nil {
				slog.Error("Failed to set palette", "err", err)
			} else {
				a.Palette = p
			}
		}
	}

	return len, nil
}

func (a *App) ROMMeta() *db.RomDto {
	program := a.DB.FindProgram(a.ROMHash)
	if program == nil {
		return nil // Unknown ROM
	}
	rom := program.ROMs[a.ROMHash]

	return &rom
}

func (a *App) ROMInfo() string {
	program := a.DB.FindProgram(a.ROMHash)
	if program == nil {
		return "Unknown"
	}
	return program.Info()
}

func (a *App) ROMConf(meta *db.RomDto, ext string) chip8.PlatformConf {
	conf := chip8.DefaultConf
	platform, ok := chip8.PlatformByExt[ext]
	if ok {
		platConf, ok := chip8.ConfByPlatform[platform]
		if ok {
			conf = platConf
		}
	}

	if meta == nil {
		slog.Info("Unknown ROM")
		return conf
	}

	for _, id := range meta.Platforms {
		if id != "megachip8" { // not supported
			platform := a.DB.FindPlatform(id)

			if platform != nil {
				slog.Info("Platform:", "id", id)
				conf.Quirks = chip8.Quirks{
					Shift:       platform.Quirks.Shift,
					MemIncIByX:  platform.Quirks.MemoryIncrementByX,
					MemLeaveI:   platform.Quirks.MemoryLeaveIUnchanged,
					Wrap:        platform.Quirks.Wrap,
					Jump:        platform.Quirks.Jump,
					VBlankWait:  platform.Quirks.VBlank,
					VFReset:     platform.Quirks.Logic,
					ScaleScroll: platform.Quirks.ScaleScroll,
				}

				if platform.DefaultTickrate > 0 {
					conf.Tickrate = platform.DefaultTickrate
				}

				break
			}
		}
	}

	if meta.Tickrate > 0 {
		conf.Tickrate = meta.Tickrate
	}

	return conf
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

	return Color{byte(ri), byte(gi), byte(bi), byte(255)}, nil
}
