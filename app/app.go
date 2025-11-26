package app

import (
	"fmt"
	"os"

	"github.com/mxmgorin/ch8go/app/db"
	"github.com/mxmgorin/ch8go/chip8"
)

type ROMInfo struct {
	Title       string
	Release     string
	Authors     []string
	Description string
	Platform    *db.PlatformDto
}

type App struct {
	Emu     chip8.Emu
	DB      *db.DB
	RomHash string
}

func NewApp() *App {
	db, err := db.NewDB()

	if err != nil {
		fmt.Println("Failed to create DB:", err)
	}

	return &App{DB: db}
}

func (a *App) LoadRom(path string) int {
	rom, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}

	a.RomHash = db.SHA1Of(rom)
	a.Emu.LoadRom(rom)

	return len(rom)
}

func (a *App) ROMInfo() *ROMInfo {
	program := a.DB.FindProgram(a.RomHash)
	if program == nil {
		return nil // Unknown ROM
	}

	romDto, ok := program.ROMs[a.RomHash]
	if !ok || len(romDto.Platforms) == 0 {
		return nil // Corrupted or incomplete DB entry
	}

	platformId := romDto.Platforms[0]

	return &ROMInfo{
		Title:       program.Title,
		Release:     program.Release,
		Authors:     program.Authors,
		Description: program.Description,
		Platform:    a.DB.FindPlatform(platformId),
	}
}
