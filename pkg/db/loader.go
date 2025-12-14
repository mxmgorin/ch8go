package db

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed programs.json platforms.json sha1-hashes.json
var fs embed.FS

func loadMeta(db *MetaDB) error {
	if err := loadJSON("programs.json", &db.programs); err != nil {
		return err
	}

	if err := loadJSON("platforms.json", &db.platforms); err != nil {
		return err
	}

	if err := loadJSON("sha1-hashes.json", &db.hashes); err != nil {
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

type ProgramMeta struct {
	Title       string             `json:"title"`
	Release     string             `json:"release"`
	Authors     []string           `json:"authors"`
	Description string             `json:"description"`
	ROMs        map[string]ROMMeta `json:"roms"`
}

func (p *ProgramMeta) Info() string {
	authors := strings.Join(p.Authors, ", ")

	return fmt.Sprintf(
		"%s\nReleased: %s\nAuthors: %s\n\n%s",
		p.Title,
		p.Release,
		authors,
		p.Description,
	)
}

type ROMMeta struct {
	File          string         `json:"file"`
	Platforms     []string       `json:"platforms"`
	Description   string         `json:"description,omitempty"`
	EmbeddedTitle string         `json:"embeddedTitle,omitempty"`
	Tickrate      int            `json:"tickrate"`
	Colors        *ROMColorsMeta `json:"colors,omitempty"`
	Keys          map[string]int `json:"keys"`
}

func (r *ROMMeta) KeysInfo() (keys string) {
	if len(r.Keys) > 0 {
		var b strings.Builder
		for k, v := range r.Keys {
			fmt.Fprintf(&b, "%s: %d\n", k, v)
		}
		keys = fmt.Sprintf("Keys:\n%s\n", b.String())
	}
	return
}

type ROMColorsMeta struct {
	Pixels  []string `json:"pixels"`  // e.g. ["#aa4400", "#ffaa00", "#ff6600", "#662200"]
	Buzzer  string   `json:"buzzer"`  // e.g. "#ffaa00"
	Silence string   `json:"silence"` // e.g. "#000000"
}

type PlatformMeta struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	Release            string     `json:"release"` // ISO date string
	DisplayResolutions []string   `json:"displayResolutions"`
	DefaultTickrate    int        `json:"defaultTickrate"`
	Quirks             QuirksMeta `json:"quirks"`
}

type QuirksMeta struct {
	Shift                 bool `json:"shift"`
	MemoryIncrementByX    bool `json:"memoryIncrementByX"`
	MemoryLeaveIUnchanged bool `json:"memoryLeaveIUnchanged"`
	Wrap                  bool `json:"wrap"`
	Jump                  bool `json:"jump"`
	VBlank                bool `json:"vblank"`
	Logic                 bool `json:"logic"`
	ScaleScroll           bool `json:"scaleScroll"`
}
