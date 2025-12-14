package host

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/mxmgorin/ch8go/pkg/chip8"
)

const (
	keyFrames = 120
	runFrames = 500_00
	frameDelta = time.Second / 60
)

var (
	outputPNG = flag.Bool("output-png", false, "write output PNG images")
	roms      = map[string]string{
		"../../testdata/roms/test/corax_test_opcode.ch8":     "663dc43daa22fa6450a29a4a799dee948b29d4f60699d1d1acf5907b030f0721",
		"../../testdata/roms/test/timendus/1-chip8-logo.ch8": "120fbab26afb931a193082a3290f3606b70ca566625b9f1067041ca2b70deaa1",
		"../../testdata/roms/test/timendus/2-ibm-logo.ch8":   "1bf96f46bc964efb0985aa88d8eeefe0229eb7c4d56ae1cd239caa9a5ea95c6c",
		"../../testdata/roms/test/timendus/3-corax+.ch8":     "af252a64884c2b1bc6772ae5b837ffcba82e083e73a8b695cf2f44b54f9529b3",
		"../../testdata/roms/test/timendus/4-flags.ch8":      "3ce4308fd0add55e5c84e036e13c2a25c3000f48120ca66d8003997a1fc77aa1",
		"../../testdata/roms/test/octo/bigfont.ch8":          "8a633131f1ac58031af7e570a113bde9f32877de24e4dd60488fe93b2dd627e5",
		"../../testdata/roms/test/octo/testbranch.ch8":       "4a28893ac197e80c0ef6342adc05a7ab95b4181b35e34da83da15ae3226e36aa",
		"../../testdata/roms/test/octo/testcollide.ch8":      "8dbf0bfcdb3f64580b7f18c03ef4c332c1398accc7932e0d44e8837d0b4dd76f",
		"../../testdata/roms/test/octo/testcompare.ch8":      "c0c45fbc3b5992cc12444ea91434f10376425822d2d5f00ff4ad198be3423178",
		"../../testdata/roms/test/octo/testquirks.ch8":       "4cb4f8029afb553bdfd5cdf3f72cdff254ad8ce4c7835ae2e8ea089e830dd9fe",
		"../../testdata/roms/test/octo/testunpack.ch8":       "1a8d5ef7e47564fbe9bebad4889881d47c597f94a654e8c58e339ad7ddaa1a23",
	}
)

func TestROMs(t *testing.T) {
	for path, expectedHash := range roms {
		t.Run(path, func(t *testing.T) {
			runROM(t, path, expectedHash)
		})
	}
}

func TestQuirksChip8(t *testing.T) {
	path := "../../testdata/roms/test/timendus/5-quirks.ch8"
	expectedHash := "9d83cd2da005411781d9e52f6b1c805e3678711098d53fbf954431e204717ce4"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksChip8)

	pressAndReleaseKey(a, 0x1)

	runAndAssert(t, path, a, expectedHash)
}

func TestQuirksSuperChipModern(t *testing.T) {
	path := "../../testdata/roms/test/timendus/5-quirks.ch8"
	expectedHash := "2b8be7cfa57527e142a321d4b6643862a4de8534e45ae7933535c0451f70b429"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(a, 0x2)
	pressAndReleaseKey(a, 0x1)

	runAndAssert(t, path, a, expectedHash)
}

func TestQuirksSuperChipLegacy(t *testing.T) {
	path := "../../testdata/roms/test/timendus/5-quirks.ch8"
	expectedHash := "b9a1bdffa9dd3ee96a97d45234f6dd79fbc80b269dae7f8ceab3a9ff8d50a083"

	key := chip8.Key2
	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChip11)

	pressAndReleaseKey(a, key)
	pressAndReleaseKey(a, key)

	runAndAssert(t, path, a, expectedHash)
}

func TestScrollSuperChipLowresLegacy(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	expectedHash := "b02eec5a6ea5042ab488f9d82ccc7262e5da140fcfaee870879f9e1fcb9ed6d5"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChip11)

	pressAndReleaseKey(a, 0x1)
	pressAndReleaseKey(a, 0x1)
	pressAndReleaseKey(a, 0x2)

	runAndAssert(t, path, a, expectedHash)
}

func TestScrollSuperChipLowresModern(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	expectedHash := "b02eec5a6ea5042ab488f9d82ccc7262e5da140fcfaee870879f9e1fcb9ed6d5"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(a, 0x1)
	pressAndReleaseKey(a, 0x1)
	pressAndReleaseKey(a, 0x1)

	runAndAssert(t, path, a, expectedHash)
}

func TestScrollSuperChipHires(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	expectedHash := "f61e02aacf428cf95fcffdca2b075606685136c627aa20ba694ad9e5cec01a8c"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(a, 0x1)
	pressAndReleaseKey(a, 0x2)

	runAndAssert(t, path, a, expectedHash)
}

func TestScrollXOChipLowres(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	expectedHash := "39e98e70eb0242da3b192accc245ff578e92edcbb359041bd05d4eb5ce7dfb05"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksXOChip)

	pressAndReleaseKey(a, 0x2)
	pressAndReleaseKey(a, 0x1)

	runAndAssert(t, path, a, expectedHash)
}

func TestScrollXOChipHires(t *testing.T) {
	path := "../../testdata/roms/test/timendus/8-scrolling.ch8"
	expectedHash := "357150e48b9338011513aa5941de1c208cb1d76b3a8ebbea4ed76d7bef80c9c3"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksXOChip)

	pressAndReleaseKey(a, 0x2)
	pressAndReleaseKey(a, 0x2)

	runAndAssert(t, path, a, expectedHash)
}

func TestKeypadDown(t *testing.T) {
	path := "../../testdata/roms/test/timendus/6-keypad.ch8"
	expectedHash := "75bb2c1659813140beefccbf2b8c23b0dff0a58acaf006b13ab148397879f109"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(a, 0x1)

	for key := range chip8.KeyCount {
		a.VM.Keypad.Press(key)
	}

	runAndAssert(t, path, a, expectedHash)
}

func TestKeypadUp(t *testing.T) {
	path := "../../testdata/roms/test/timendus/6-keypad.ch8"
	expectedHash := "dfb2a361a04cb47336ab97509eef04b555940455018216ea98c485a3db17c6e0"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(a, 0x2)

	for key := range chip8.KeyCount {
		a.VM.Keypad.Press(key)
	}

	runAndAssert(t, path, a, expectedHash)
}

func TestKeypadGetkey(t *testing.T) {
	path := "../../testdata/roms/test/timendus/6-keypad.ch8"
	expectedHash := "9bef3db51e363351faac792cffbce35477686bdb2679e7d03a6b8f5999d9ab20"

	a := setup(t, path)
	a.VM.SetQuirks(chip8.QuirksSChipModern)

	pressAndReleaseKey(a, 0x3)
	pressAndReleaseKey(a, 0x3)

	runAndAssert(t, path, a, expectedHash)
}

func pressAndReleaseKey(app *Emu, key chip8.Key) {
	app.VM.Keypad.Press(key)

	for range keyFrames {
		app.runFrame(frameDelta)
	}

	app.VM.Keypad.Release(key)

	for range keyFrames {
		app.runFrame(frameDelta)
	}
}

func runROM(t *testing.T, path, expectedHash string) {
	t.Helper() // marks this as a test helper

	a := setup(t, path)

	runAndAssert(t, path, a, expectedHash)
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

func runAndAssert(t *testing.T, path string, app *Emu, expected string) {
	t.Helper()

	for range runFrames {
		app.runFrame(frameDelta)
	}

	fb := app.RunFrame()
	actual := fb.Hash()

	if *outputPNG {
		prefix := expected[:6]
		out := filepath.Join(
			"..",
			"testdata",
			"output",
			filepath.Base(path)+"_"+prefix+".png",
		)
		if err := fb.SavePNG(out); err != nil {
			t.Fatal(err)
		}
	}

	if actual != expected {
		t.Fatalf("hash mismatch for %s:\nexpected: %s\nactual: %s",
			path, expected, actual)
	}
}
