package db

import (
	"strings"
	"testing"
)

func TestSHA1Of(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"abc", "a9993e364706816aba3e25717850c26c9cd0d89d"},
	}
	for _, tt := range tests {
		if got := SHA1Of([]byte(tt.in)); got != tt.want {
			t.Errorf("SHA1Of(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNewMetaDB(t *testing.T) {
	db, err := NewMetaDB()
	if err != nil {
		t.Fatalf("NewMetaDB() error = %v", err)
	}
	if len(db.programs) == 0 {
		t.Error("expected non-empty programs")
	}
	if len(db.platforms) == 0 {
		t.Error("expected non-empty platforms")
	}
	if len(db.hashes) == 0 {
		t.Error("expected non-empty hashes")
	}
}

func TestPlatformLookup(t *testing.T) {
	db, err := NewMetaDB()
	if err != nil {
		t.Fatalf("NewMetaDB() error = %v", err)
	}

	want := db.platforms[0]
	got := db.Platform(want.ID)
	if got == nil {
		t.Fatalf("Platform(%q) = nil, want non-nil", want.ID)
	}
	if got.ID != want.ID {
		t.Errorf("Platform(%q).ID = %q, want %q", want.ID, got.ID, want.ID)
	}

	if p := db.Platform("no-such-platform"); p != nil {
		t.Errorf("Platform(unknown) = %+v, want nil", p)
	}
}

func TestProgramAndROMLookup(t *testing.T) {
	db, err := NewMetaDB()
	if err != nil {
		t.Fatalf("NewMetaDB() error = %v", err)
	}

	// Grab any known hash from the embedded database.
	var hash string
	for h := range db.hashes {
		hash = h
		break
	}
	if hash == "" {
		t.Fatal("no hashes in database")
	}

	if p := db.Program(hash); p == nil {
		t.Errorf("Program(%q) = nil, want non-nil", hash)
	}
	if r := db.ROM(hash); r == nil {
		t.Errorf("ROM(%q) = nil, want non-nil", hash)
	}

	const unknown = "0000000000000000000000000000000000000000"
	if p := db.Program(unknown); p != nil {
		t.Errorf("Program(unknown) = %+v, want nil", p)
	}
	if r := db.ROM(unknown); r != nil {
		t.Errorf("ROM(unknown) = %+v, want nil", r)
	}
}

func TestProgramMetaInfo(t *testing.T) {
	p := ProgramMeta{
		Title:       "Test Game",
		Release:     "2020",
		Authors:     []string{"Alice", "Bob"},
		Description: "A demo.",
	}
	info := p.Info()
	for _, want := range []string{"Test Game", "2020", "Alice, Bob", "A demo."} {
		if !strings.Contains(info, want) {
			t.Errorf("Info() = %q, missing %q", info, want)
		}
	}
}

func TestROMMetaKeysInfo(t *testing.T) {
	empty := ROMMeta{}
	if got := empty.KeysInfo(); got != "" {
		t.Errorf("KeysInfo() with no keys = %q, want empty", got)
	}

	r := ROMMeta{Keys: map[string]int{"up": 5}}
	if got := r.KeysInfo(); !strings.Contains(got, "up: 5") {
		t.Errorf("KeysInfo() = %q, want to contain %q", got, "up: 5")
	}
}
