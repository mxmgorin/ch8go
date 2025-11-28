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

func TestROMs(t *testing.T) {
	for path, expectedHash := range roms {
		t.Run(path, func(t *testing.T) {
			testROM(t, path, expectedHash)
		})
	}
}

func TestQuirksChip8(t *testing.T) {
	path := "../roms/test/timendus/5-quirks.ch8"
	expectedHash := "cfc94dc6acf6f832242372429c1a89f29dd715e19c8122efb83363a16f873146"

	vm := loadVM(t, path)
	vm.SetQuirks(chip8.QuirksOriginalChip8)

	pressAndReleaseKey(vm, 0x1)

	assert(t, path, vm, expectedHash)
}

func TestQuirksSuperChipModern(t *testing.T) {
	path := "../roms/test/timendus/5-quirks.ch8"
	expectedHash := "064622ef39ecc953146f83e6ff7b8d954ede3dd3e21a551afca175cf2cda8f99"

	vm := loadVM(t, path)
	vm.SetQuirks(chip8.QuirksSuperChip11)

	pressAndReleaseKey(vm, 0x2)
	pressAndReleaseKey(vm, 0x1)

	assert(t, path, vm, expectedHash)
}

func TestQuirksSuperChipLegacy(t *testing.T) {
	path := "../roms/test/timendus/5-quirks.ch8"
	expectedHash := "a754278cf652b6018568766247915b6034412a00632a90e0e06770421f3a501a"

	key := byte(0x2)
	vm := loadVM(t, path)
	vm.SetQuirks(chip8.QuirksSuperChipLegacy)

	pressAndReleaseKey(vm, key)
	pressAndReleaseKey(vm, key)

	assert(t, path, vm, expectedHash)
}

func TestScrollSuperChipHires(t *testing.T) {
	path := "../roms/test/timendus/8-scrolling.ch8"
	expectedHash := "085d7d83b14b56618323684a700efeeb85ddc8e2f1184a1a7467e23675173019"

	vm := loadVM(t, path)
	vm.SetQuirks(chip8.QuirksSuperChip11)

	pressAndReleaseKey(vm, 0x1)
	pressAndReleaseKey(vm, 0x2)

	assert(t, path, vm, expectedHash)
}

func pressAndReleaseKey(vm *chip8.VM, key byte) {
	frames := 6000
	dt := 0.016
	vm.Keypad.Press(key)

	for range frames {
		vm.RunFrame(dt)
	}

	vm.Keypad.Release(key)

	for range frames {
		vm.RunFrame(dt)
	}
}

func testROM(t *testing.T, path, expectedHash string) {
	t.Helper() // marks this as a test helper

	vm := loadVM(t, path)

	assert(t, path, vm, expectedHash)
}

func loadVM(t *testing.T, path string) *chip8.VM {
	t.Helper() // marks this as test helper

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read ROM %s: %v", path, err)
	}

	vm := chip8.NewVM()
	vm.LoadROM(data)

	return vm
}

func assert(t *testing.T, path string, vm *chip8.VM, expected string) {
	t.Helper()

	for range 600_000 {
		vm.RunFrame(0.016)
	}

	actual := hash(vm.Display.Pixels[:])

	if actual != expected {
		t.Fatalf("hash mismatch for %s:\nexpected: %s\nactual: %s",
			path, expected, actual)
	}
}

func hash(buff []byte) string {
	sum := sha256.Sum256(buff)
	return fmt.Sprintf("%x", sum[:])
}
