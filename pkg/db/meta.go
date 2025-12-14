package db

import (
	"crypto/sha1"
	"fmt"
)

type MetaDB struct {
	programs  []ProgramMeta
	platforms []PlatformMeta
	hashes    map[string]int
}

func NewMetaDB() (*MetaDB, error) {
	db := MetaDB{}
	if err := loadMeta(&db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *MetaDB) Platform(id string) (platform *PlatformMeta) {
	for i := range db.platforms {
		if db.platforms[i].ID == id {
			platform = &db.platforms[i]
			break
		}
	}

	return platform
}

func (db *MetaDB) ROM(hash string) *ROMMeta {
	program := db.Program(hash)
	if program == nil {
		return nil // Unknown ROM
	}
	rom := program.ROMs[hash]

	return &rom
}

func (db *MetaDB) Program(hash string) *ProgramMeta {
	idx, ok := db.hashes[hash]
	if !ok {
		return nil
	}

	if idx < 0 || idx >= len(db.programs) {
		return nil
	}

	return &db.programs[idx]
}

func SHA1Of(data []byte) string {
	sum := sha1.Sum(data)
	return fmt.Sprintf("%x", sum)
}
