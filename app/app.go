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
	VM      chip8.VM
	DB      *db.DB
	ROMHash string
}

func NewApp() *App {
	db, err := db.NewDB()

	if err != nil {
		fmt.Println("Failed to create DB:", err)
	}

	return &App{DB: db, VM: *chip8.NewVM()}
}

func (a *App) HasROM() bool {
	return a.ROMHash != ""
}

func (a *App) ReadROM(path string) int {
	rom, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}

	return a.LoadROM(rom)
}

func (a *App) LoadROM(rom []byte) int {
	a.ROMHash = db.SHA1Of(rom)
	a.VM.LoadROM(rom)

	return len(rom)
}

func (a *App) ROMInfo() *db.RomDto {
	program := a.DB.FindProgram(a.ROMHash)
	if program == nil {
		return nil // Unknown ROM
	}
	rom := program.ROMs[a.ROMHash]

	return &rom
}
