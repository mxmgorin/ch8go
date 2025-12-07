package db

import (
	"crypto/sha1"
	"fmt"
)

type DB struct {
	Programs  []ProgramDto
	Platforms []PlatformDto
	Hashes    map[string]int
}

func NewDB() (*DB, error) {
	db := DB{}
	if err := load(&db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *DB) FindPlatform(id string) (platform *PlatformDto) {
	for i := range db.Platforms {
		if db.Platforms[i].ID == id {
			platform = &db.Platforms[i]
			break
		}
	}

	return platform
}

func (db *DB) FindProgram(hash string) *ProgramDto {
	idx, ok := db.Hashes[hash]
	if !ok {
		return nil
	}

	if idx < 0 || idx >= len(db.Programs) {
		return nil
	}
	return &db.Programs[idx]
}

func SHA1Of(data []byte) string {
	sum := sha1.Sum(data)
	return fmt.Sprintf("%x", sum)
}
