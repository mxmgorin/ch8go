package tests

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/mxmgorin/ch8go/chip8"
)

var roms = map[string]string{
	"../roms/test/corax_test_opcode.ch8":     "53e3214ea1cfc7c81f8d5f05c652d5196e8a547fa9d762e65370462fce08bd24",
	"../roms/test/timendus/1-chip8-logo.ch8": "081dcfdcf4fb9384f2c6a9161ddce5349efe88939415dc721ce5c745e538ba81",
	"../roms/test/timendus/2-ibm-logo.ch8":   "daed3448a74f730cf1867672fe07073baf63bf01cad5ba25c93d996ee0ad33c2",
	"../roms/test/timendus/3-corax+.ch8":     "91eab06fdca9acf793593135d551fafd5c0e3764135c9e066ad526c67563f5fa",
	"../roms/test/timendus/4-flags.ch8":      "7b6d24ec24c5cd2b7ed4b2176a6a28f8b3b6120d4d1503190674901063113556",
}

func TestRoms(t *testing.T) {
	for path, expectedHash := range roms {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read ROM: %v", err)
			}

			vm := chip8.NewVM()
			vm.LoadROM(data)

			for range 1_000_000 {
				vm.Step()
			}

			actualHash := hash(vm.Display.Pixels[:])

			if actualHash != expectedHash {
				t.Fatalf("hash mismatch:\nexpected: %s\nactual: %s", expectedHash, actualHash)
			}
		})
	}
}

func hash(buff []byte) string {
	sum := sha256.Sum256(buff)
	return fmt.Sprintf("%x", sum[:])
}
