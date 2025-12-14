package host

import (
	"bytes"
	"errors"
	"flag"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

const (
	keyFrames  = 120
	runFrames  = 500_00
	frameDelta = time.Second / 60
)

var (
	updateGolden = flag.Bool("update-golden", false, "write golden PNG images")
	romPaths     = []string{
		"../../testdata/roms/test/corax_test_opcode.ch8",
		"../../testdata/roms/test/timendus/1-chip8-logo.ch8",
		"../../testdata/roms/test/timendus/2-ibm-logo.ch8",
		"../../testdata/roms/test/timendus/3-corax+.ch8",
		"../../testdata/roms/test/timendus/4-flags.ch8",
		"../../testdata/roms/test/octo/bigfont.ch8",
		"../../testdata/roms/test/octo/testbranch.ch8",
		"../../testdata/roms/test/octo/testcollide.ch8",
		"../../testdata/roms/test/octo/testcompare.ch8",
		"../../testdata/roms/test/octo/testquirks.ch8",
		"../../testdata/roms/test/octo/testunpack.ch8",
	}
)

func TestROMs(t *testing.T) {
	for _, path := range romPaths {
		t.Run(path, func(t *testing.T) {
			runROM(t, path, "")
		})
	}
}

func TestQuirksChip8(t *testing.T) {
	path := "../../testdata/roms/test/timendus/5-quirks.ch8"
	name := "chip8"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksChip8)

	pressAndReleaseKey(emu, 0x1)

	runAndAssert(t, path, emu, name)
}

func TestQuirksSChipModern(t *testing.T) {
	path := "../../testdata/roms/test/timendus/5-quirks.ch8"
	name := "schip-modern"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(emu, 0x2)
	pressAndReleaseKey(emu, 0x1)

	runAndAssert(t, path, emu, name)
}

func TestQuirksSChipLegacy(t *testing.T) {
	path := "../../testdata/roms/test/timendus/5-quirks.ch8"
	name := "schip-legacy"

	key := chip8.Key2
	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChip11)

	pressAndReleaseKey(emu, key)
	pressAndReleaseKey(emu, key)

	runAndAssert(t, path, emu, name)
}

func TestScrollSChipLowresLegacy(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	name := "schip-lowres-legacy"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChip11)

	pressAndReleaseKey(emu, 0x1)
	pressAndReleaseKey(emu, 0x1)
	pressAndReleaseKey(emu, 0x2)

	runAndAssert(t, path, emu, name)
}

func TestScrollSChipLowresModern(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	name := "schip-lowres-modern"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(emu, 0x1)
	pressAndReleaseKey(emu, 0x1)
	pressAndReleaseKey(emu, 0x1)

	runAndAssert(t, path, emu, name)
}

func TestScrollSChipHiresModern(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	name := "schip-hires-modern"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(emu, 0x1)
	pressAndReleaseKey(emu, 0x2)

	runAndAssert(t, path, emu, name)
}

func TestScrollXOChipLowres(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	name := "xo-chip-lowres"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksXOChip)

	pressAndReleaseKey(emu, 0x2)
	pressAndReleaseKey(emu, 0x1)

	runAndAssert(t, path, emu, name)
}

func TestScrollXOChipHires(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	name := "xo-chip-hires"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksXOChip)

	pressAndReleaseKey(emu, 0x2)
	pressAndReleaseKey(emu, 0x2)

	runAndAssert(t, path, emu, name)
}

func TestKeypadDown(t *testing.T) {
	path := "../../testdata/roms/test/timendus/6-keypad.ch8"
	name := "key-down"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(emu, 0x1)

	for key := range chip8.KeyCount {
		emu.VM.Keypad.Press(key)
	}

	runAndAssert(t, path, emu, name)
}

func TestKeypadUp(t *testing.T) {
	path := "../../testdata/roms/test/timendus/6-keypad.ch8"
	name := "key-up"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(emu, 0x2)

	for key := range chip8.KeyCount {
		emu.VM.Keypad.Press(key)
	}

	runAndAssert(t, path, emu, name)
}

func TestKeypadGetkey(t *testing.T) {
	path := "../../testdata/roms/test/timendus/6-keypad.ch8"
	name := "get-key"

	emu := setup(t, path)
	emu.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(emu, 0x3)
	pressAndReleaseKey(emu, 0x3)

	runAndAssert(t, path, emu, name)
}

func pressAndReleaseKey(emu *Emu, key chip8.Key) {
	emu.VM.Keypad.Press(key)

	for range keyFrames {
		emu.runFrame(frameDelta)
	}

	emu.VM.Keypad.Release(key)

	for range keyFrames {
		emu.runFrame(frameDelta)
	}
}

func runROM(t *testing.T, path, prefix string) {
	t.Helper() // marks this as a test helper

	emu := setup(t, path)

	runAndAssert(t, path, emu, prefix)
}

func setup(t *testing.T, path string) *Emu {
	t.Helper() // marks this as test helper

	app, err := NewEmu()
	if err != nil {
		t.Fatalf("failed NewApp: %v", err)
	}

	if _, err := app.ReadROM(path); err != nil {
		t.Fatalf("failed to read ROM %s: %v", path, err)
	}

	return app
}

func runAndAssert(t *testing.T, romPath string, emu *Emu, suffix string) {
	t.Helper()

	for range runFrames {
		emu.runFrame(frameDelta)
	}

	fb := emu.RunFrame()

	goldenPath := goldenPath(romPath, suffix)

	if *updateGolden {
		if err := fb.SavePNG(goldenPath); err != nil {
			t.Fatal(err)
		}
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}

	got, err := fb.PNG()
	if err := comparePNG(got, want); err != nil {
		t.Fatalf("framebuffer mismatch for %s: %v", romPath, err)
	}
}

func comparePNG(a, b []byte) error {
	imgA, err := png.Decode(bytes.NewReader(a))
	if err != nil {
		return err
	}
	imgB, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(imgA, imgB) {
		return errors.New("pixel mismatch")
	}
	return nil
}

func goldenPath(romPath, suffix string) string {
	if suffix != "" {
		suffix = "_" + suffix
	}
	base := filepath.Base(romPath)
	filename := strings.TrimSuffix(base, filepath.Ext(base)) + suffix + ".png"
	path := filepath.Join(
		"..",
		"..",
		"testdata",
		"golden",
		filename,
	)
	return path
}
