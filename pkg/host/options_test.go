package host

import (
	"flag"
	"io"
	"testing"
)

func TestValidateROMPath(t *testing.T) {
	if err := (&Options{}).ValidateROMPath(); err == nil {
		t.Error("empty ROM path should be invalid")
	}
	if err := (&Options{ROMPath: "game.ch8"}).ValidateROMPath(); err != nil {
		t.Errorf("valid ROM path returned error: %v", err)
	}
}

func TestParseOptions(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	opts, err := ParseOptions(fs, []string{"--rom", "game.ch8", "--scale", "5"})
	if err != nil {
		t.Fatalf("ParseOptions error = %v", err)
	}
	if opts.ROMPath != "game.ch8" || opts.Scale != 5 {
		t.Errorf("ParseOptions = %+v, want {game.ch8 5}", opts)
	}

	// Defaults when no flags are provided.
	def := flag.NewFlagSet("def", flag.ContinueOnError)
	opts, err = ParseOptions(def, nil)
	if err != nil {
		t.Fatalf("ParseOptions defaults error = %v", err)
	}
	if opts.Scale != 12 || opts.ROMPath != "" {
		t.Errorf("defaults = %+v, want {\"\" 12}", opts)
	}

	// An unknown flag surfaces a parse error.
	bad := flag.NewFlagSet("bad", flag.ContinueOnError)
	bad.SetOutput(io.Discard)
	if _, err := ParseOptions(bad, []string{"--nope"}); err == nil {
		t.Error("unknown flag should produce an error")
	}
}
