package host

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/mxmgorin/ch8go/pkg/chip8"
	"github.com/mxmgorin/ch8go/pkg/db"
)

type Emu struct {
	VM            *chip8.VM
	MetaDB        *db.MetaDB
	ROMHash       string
	Palette       Palette
	Paused        bool
	FrameBuffer   FrameBuffer
	lastFrameTime time.Time
}

func NewEmu() (*Emu, error) {
	metaDB, err := db.NewMetaDB()

	if err != nil {
		slog.Error("Failed to create MetaDB", "err", err)
	}

	vm := chip8.NewVM()
	size := vm.Display.Size()

	return &Emu{
		MetaDB:        metaDB,
		VM:            vm,
		FrameBuffer:   newFrameBuffer(size.Width, size.Height, 4),
		lastFrameTime: time.Now(),
		Palette:       DefaultPalette,
	}, nil
}

func (e *Emu) Loaded() bool {
	return e.ROMHash != ""
}

func (e *Emu) ReadROM(path string) (int, error) {
	rom, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return e.LoadROM(rom, filepath.Ext(path))
}

func (e *Emu) LoadROM(rom []byte, ext string) (int, error) {
	e.Palette = DefaultPalette
	e.ROMHash = db.SHA1Of(rom)
	len := len(rom)

	slog.Info("ROM loaded:", "size", len, "hash", e.ROMHash, "ext", ext)

	if err := e.VM.LoadROM(rom); err != nil {
		return 0, err
	}

	rm := e.ROMMeta()
	rc := e.ROMConf(rm, ext)
	e.VM.SetConf(rc)

	if rm != nil {
		colors := rm.Colors
		if colors != nil && colors.Pixels != nil {
			p, err := NewPalette(colors.Pixels, colors.Buzzer, colors.Silence)
			if err != nil {
				slog.Error("Failed to set palette", "err", err)
			} else {
				e.Palette = p
			}
		}
	}

	return len, nil
}

func (e *Emu) ROMMeta() *db.ROMMeta {
	return e.MetaDB.ROM(e.ROMHash)
}

func (e *Emu) ROMInfo() string {
	program := e.MetaDB.Program(e.ROMHash)
	if program == nil {
		return "Unknown"
	}
	return program.Info()
}

func (e *Emu) ROMConf(meta *db.ROMMeta, ext string) chip8.PlatformConf {
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
			platform := e.MetaDB.Platform(id)

			if platform != nil {
				slog.Info("Platform:", "id", id)
				conf.Quirks = chip8.Quirks{
					Shift:       platform.Quirks.Shift,
					MemIncIByX:  platform.Quirks.MemoryIncrementByX,
					MemLeaveI:   platform.Quirks.MemoryLeaveIUnchanged,
					Wrap:        platform.Quirks.Wrap,
					Jump:        platform.Quirks.Jump,
					WaitVBlank:  platform.Quirks.VBlank,
					ResetFlag:   platform.Quirks.Logic,
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

func (e *Emu) RunFrame() *FrameBuffer {
	now := time.Now()
	frameDelta := now.Sub(e.lastFrameTime)
	e.lastFrameTime = now

	return e.runFrame(frameDelta)
}

func (e *Emu) runFrame(frameDelta time.Duration) *FrameBuffer {
	if e.Loaded() && !e.Paused {
		state := e.VM.RunFrame(frameDelta)
		e.FrameBuffer.Update(state, &e.Palette, &e.VM.Display)
	}
	return &e.FrameBuffer
}

func ParseFlags() (string, int) {
	romPath := flag.String("rom", "", "path to CHIP-8 ROM")
	scale := flag.Int("scale", 12, "window scale")
	flag.Parse()

	if *romPath == "" {
		log.Fatal("You must provide a ROM: --rom path/to/file.ch8")
	}

	return *romPath, *scale
}
