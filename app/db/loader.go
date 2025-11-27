package db

import (
	"embed"
	"encoding/json"
)

//go:embed programs.json platforms.json sha1-hashes.json
var fs embed.FS

func load(db *DB) error {
	if err := loadJSON("programs.json", &db.Programs); err != nil {
		return err
	}

	if err := loadJSON("platforms.json", &db.Platforms); err != nil {
		return err
	}

	if err := loadJSON("sha1-hashes.json", &db.Hashes); err != nil {
		return err
	}

	return nil
}

func loadJSON(name string, v any) error {
	data, err := fs.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type ProgramDto struct {
	Title       string            `json:"title"`
	Release     string            `json:"release"`
	Authors     []string          `json:"authors"`
	Description string            `json:"description"`
	ROMs        map[string]RomDto `json:"roms"`
}

type RomDto struct {
	File          string   `json:"file"`
	Platforms     []string `json:"platforms"`
	Description   string   `json:"description,omitempty"`
	EmbeddedTitle string   `json:"embeddedTitle,omitempty"`
	Tickrate    int       `json:"tickrate"`
}

type PlatformDto struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Release            string    `json:"release"` // ISO date string
	DisplayResolutions []string  `json:"displayResolutions"`
	DefaultTickrate    int       `json:"defaultTickrate"`
	Quirks             QuirksDto `json:"quirks"`
}

type QuirksDto struct {
	Shift                 bool `json:"shift"`
	MemoryIncrementByX    bool `json:"memoryIncrementByX"`
	MemoryLeaveIUnchanged bool `json:"memoryLeaveIUnchanged"`
	Wrap                  bool `json:"wrap"`
	Jump                  bool `json:"jump"`
	VBlank                bool `json:"vblank"`
	Logic                 bool `json:"logic"`
}
