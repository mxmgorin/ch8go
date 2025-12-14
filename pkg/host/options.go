package host

import (
	"flag"
	"fmt"
)

// Options contains command-line configuration used to run the emulator on the host system.
type Options struct {
	ROMPath string
	Scale   int
}

func (o *Options) ValidateROMPath() error {
	if o.ROMPath == "" {
		return fmt.Errorf("missing required --rom flag")
	}
	return nil
}

// ParseOptions parses emulator-related command-line flags from args.
func ParseOptions(fs *flag.FlagSet, args []string) (Options, error) {
	opts := Options{}

	fs.StringVar(&opts.ROMPath, "rom", "", "path to CHIP-8 ROM")
	fs.IntVar(&opts.Scale, "scale", 12, "window scale")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}

	return opts, nil
}
