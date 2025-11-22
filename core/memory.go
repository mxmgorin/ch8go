package core

import (
	"fmt"
)

const MemorySize = 4096
const ProgramStart = 0x200

type Memory struct {
    bytes [MemorySize]byte
}

func NewMemory() *Memory {
    m := &Memory{}
    return m
}

func (m *Memory) LoadRom(bytes []byte) error {
    if len(bytes)+ProgramStart > MemorySize {
        return fmt.Errorf("ROM too large")
    }

    copy(m.bytes[ProgramStart:], bytes)
    return nil
}

func (m *Memory) Read(addr uint16) byte {
    return m.bytes[addr]
}

func (m *Memory) Write(addr uint16, val byte) {
    m.bytes[addr] = val
}
