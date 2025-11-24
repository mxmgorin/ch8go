package tests

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/mxmgorin/ch8go/chip8"
)

var roms = map[string]string{
	"../roms/test/timendus/1-chip8-logo.ch8": "aec99a55453e70020320e8c93bf582b561df516c4efb05d10f43eda1ee3c6b53",
	"../roms/test/timendus/2-ibm-logo.ch8":   "b12be07c247d2a94b678808638361649ca8d14c22f1e1468c9849fd0aefa4421",
	"../roms/test/timendus/3-corax+.ch8":     "f7accf00a65c264fadfd94280d57f6c6564115df4b99316395e8253ff1729024",
	// "../roms/test/timendus/4-flags.ch8":      "", //todo: FIXME
}

func TestRoms(t *testing.T) {
	for path, expectedHash := range roms {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read ROM: %v", err)
			}

			emu := chip8.NewEmu()
			emu.LoadRom(data)

			for range 1_000_000 {
				emu.Step()
			}

			actualHash := hash(emu.Display.Pixels[:])

			if actualHash != expectedHash {
				t.Fatalf("Buffer hash mismatch:\nExpected: %s\nActual: %s", expectedHash, actualHash)
			}
		})
	}
}

func hash(buff []byte) string {
	sum := sha256.Sum256(buff)
	return fmt.Sprintf("%x", sum[:])
}
