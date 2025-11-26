package chip8

type Quirks struct {
	// On most systems the shift opcodes take `vY` as input and stores the shifted version of `vY` into `vX`.
	// The interpreters for the HP48 took `vX` as both the input and the output, introducing the shift quirk.
	// True: `8XY6` and `8XYE` take `vX` as both input and output
	// False: `8XY6` and `8XYE` take `vY` as input and `vX` as output
	Shift bool

	// True: `FX55` and `FX65` increment the `i` register with `X`
	// False: `FX55` and `FX65` increment the `i` register with `X + 1`
	MemIncIByX bool

	// True: `FX55` and `FX65` leave the `i` register unchanged
	// False: `FX55` and `FX65` increment the `i` register
	MemIUnchanged bool

	// True: `DXYN` wraps around to the other side of the screen when drawing at the edges
	// False: `DXYN` clips when drawing at the edges of the screen
	Wrap bool

	// The jump to `<address> + v0` opcode was wronly implemented on all the HP48 interpreters as jump to `<address> + vX`, introducing the jump quirk."
	// True: `BXNN` jumps to address `XNN + vX`
	// False: `BNNN` jumps to address `NNN + v0`
	Jump bool

	// The original Cosmac VIP interpreter would wait for vertical blank before each sprite draw.
	// This was done to prevent sprite tearing on the display, but it would also act as an accidental limit on the execution speed of the program.
	// Some programs rely on this speed limit to be playable. Vertical blank happens at 60Hz, and as such its logic be combined with the timers.
	// True: `DXYN` waits for vertical blank (so max 60 sprites drawn per second)
	// False: `DXYN` draws immediately (number of sprites drawn per second only limited to number of CPU cycles per frame)
	VBlank bool

	// On the original Cosmac VIP interpreter, `vF` would be reset after each opcode that would invoke the maths coprocessor.
	// Later interpreters have not copied this behaviour.
	// True: `8XY1`, `8XY2` and `8XY3` (OR, AND and XOR) will set `vF` to zero after execution (even if `vF` is the parameter `X`)
	// False: `8XY1`, `8XY2` and `8XY3` (OR, AND and XOR) will leave `vF` unchanged (unless `vF` is the parameter `X`)
	VFReset bool
}

var (
	QuirksOriginalChip = Quirks{
		Shift:         false,
		MemIncIByX:    false,
		MemIUnchanged: false,
		Wrap:          false,
		Jump:          false,
		VBlank:        true,
		VFReset:       true,
	}
	QuirksModernChip = Quirks{
		Shift:         false,
		MemIncIByX:    false,
		MemIUnchanged: false,
		Wrap:          false,
		Jump:          false,
		VBlank:        true,
		VFReset:       true,
	}
	QuirksSchip11 = Quirks{
		Shift:         true,
		MemIUnchanged: true,
		Wrap:          false,
		Jump:          true,
		VBlank:        false,
		VFReset:       false,
	}
)
