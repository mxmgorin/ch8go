package chip8

const (
	PlatformChip8   Platform = "ch8"
	PlatformSChip11 Platform = "sc8"
	PlatformXOChip  Platform = "xo8"
)

var Platforms = map[Platform]PlatformConfig{
	PlatformChip8:   PlatformConfig{Quirks: QuirksChip8, TickRate: 15},
	PlatformSChip11: PlatformConfig{Quirks: QuirksSChip11, TickRate: 30},
	PlatformXOChip:  PlatformConfig{Quirks: QuirksXOChip, TickRate: 100},
}
var (
	// CHIP-8 was first designed by Joseph Weisbecker for the Cosmac VIP hobbyist DIY computer in 1977.
	// After publishing about the virtual instruction set in the december 1978 issue of Byte magazine it took off on more hobbyist computers.
	// One of the biggest advantages of programming in CHIP-8, apart from being relatively easy to use,
	// was the fact that CHIP-8 ROMs were binary compatible between several different hobbyist computers.
	QuirksChip8 = Quirks{
		Shift:      false,
		MemIncIByX: false,
		MemLeaveI:  false,
		Wrap:       false,
		Jump:       false,
		VBlankWait: true,
		VFReset:    true,
	}

	// This is the way CHIP-8 is usually implemented in modern times. People often don't bother implementing the vBlank quirk, which leads to a more fluid,
	// slightly faster execution. The vF reset on logic operations is also usually ignored because the impact is minimal and the quirk is fairly unknown.
	// Some ROMs have come to depend on this \"simpler\" implementation, and as a result do not run very well on the original interpreter.
	QuirksChip8Modern = Quirks{
		Shift:      false,
		MemIncIByX: false,
		MemLeaveI:  false,
		Wrap:       false,
		Jump:       false,
		VBlankWait: false,
		VFReset:    false,
	}

	// Superchip 1.1 is the platform that most \"superchip\" interpreters implement, because it is the latest version and also
	// because the difference between Superchip version 1.0 and 1.1 is pretty small.
	// This version is faster than its predecessor and adds scroll instructions and a large numeric font.
	// It does however introduces a new quirk by not incrementing the index register when reading or writing registers to memory.
	QuirksSChip11 = Quirks{
		Shift:       true,
		MemIncIByX:  false,
		MemLeaveI:   true,
		Wrap:        false,
		Jump:        true,
		VBlankWait:  true,
		VFReset:     false,
		ScaleScroll: false,
	}

	QuirksSChipModern = Quirks{
		Shift:       true,
		MemIncIByX:  false,
		MemLeaveI:   true,
		Wrap:        false,
		Jump:        true,
		VBlankWait:  false,
		VFReset:     false,
		ScaleScroll: true,
	}

	QuirksXOChip = Quirks{
		Shift:       false,
		MemIncIByX:  false,
		MemLeaveI:   false,
		Wrap:        true,
		Jump:        false,
		VBlankWait:  false,
		VFReset:     false,
		ScaleScroll: true,
	}
)

type Platform string

type PlatformConfig struct {
	Quirks   Quirks
	TickRate int
}

func (c *PlatformConfig) CPUHz() float64 {
	return float64(c.TickRate) * 60.0
}

type Quirks struct {
	// Shift controls how the shift opcodes interpret their input and output
	// registers.
	//
	// On most systems, the shift opcodes (`8XY6` and `8XYE`) take `vY` as input
	// and store the shifted result into `vX`.
	//
	// The HP-48 interpreters behaved differently: they used `vX` as both the input
	// and the output register. This behavior is known as the shift quirk.
	//
	// When true:
	//   - `8XY6` and `8XYE` take `vX` as both the input and the output.
	//
	// When false:
	//   - `8XY6` and `8XYE` take `vY` as input and store the result in `vX`.
	Shift bool

	// MemIncIByX controls the first load/store quirk.
	//
	// On most systems, storing or loading registers with `FX55` and `FX65`
	// increments the `I` register by `X + 1`, i.e., once for each register
	// transferred. The CHIP-48 interpreter for the HP-48 incremented `I` only
	// by `X`, introducing this quirk.
	//
	// When true:
	//   - `FX55` and `FX65` increment `I` by `X`.
	//
	// When false:
	//   - `FX55` and `FX65` increment `I` by `X + 1`.
	MemIncIByX bool

	// MemLeaveI controls the second load/store quirk.
	//
	// On most systems, `FX55` and `FX65` increment the `I` register based on
	// how many registers are read or written. The Superchip 1.1 interpreter for
	// the HP-48 did not increment `I` at all, introducing this quirk.
	//
	// When true:
	//   - `FX55` and `FX65` leave the `I` register unchanged.
	//
	// When false:
	//   - `FX55` and `FX65` increment the `I` register normally.
	MemLeaveI bool

	// Wrap controls how sprites behave when drawn at the edges of the screen.
	//
	// Most CHIP-8 systems clip sprites when drawing near screen boundaries.
	// The Octo interpreter—later used in the XO-CHIP variant—wraps sprites
	// around to the opposite side of the screen instead, introducing the wrap
	// quirk.
	//
	// When true:
	//   - `DXYN` wraps sprites around the screen edges.
	//
	// When false:
	//   - `DXYN` clips sprites at the screen boundaries.
	Wrap bool

	// Jump controls how the indexed jump opcode interprets its offset register.
	//
	// The jump-to-<address+V0> opcode (`BNNN`) was incorrectly implemented on all
	// HP-48 interpreters as <address+Vx>, introducing what is known as the
	// jump quirk.
	//
	// When true:
	//   - `BXNN` jumps to address `XNN + Vx`.
	//
	// When false:
	//   - `BNNN` jumps to address `NNN + V0`.
	Jump bool

	// The original Cosmac VIP interpreter would wait for vertical blank before each sprite draw.
	// This was done to prevent sprite tearing on the display, but it would also act as an accidental limit on the execution speed of the program.
	// Some programs rely on this speed limit to be playable. Vertical blank happens at 60Hz, and as such its logic be combined with the timers.
	// True: `DXYN` waits for vertical blank (so max 60 sprites drawn per second).
	// False: `DXYN` draws immediately (number of sprites drawn per second only limited to number of CPU cycles per frame).
	VBlankWait bool

	// VFReset controls how the emulator handles the `vF` register after logic
	// operations.
	//
	// On the original COSMAC VIP interpreter, `vF` was reset after each opcode
	// that invoked the math coprocessor.
	//
	// Later interpreters did not copy this behavior.
	//
	// When true:
	//   - `8XY1`, `8XY2`, and `8XY3` (OR, AND, XOR) will set `vF` to zero after
	//     execution, even if `vF` is the X register.
	//
	// When false:
	//   - `8XY1`, `8XY2`, and `8XY3` will leave `vF` unchanged, unless `vF` is
	//     the X register.
	VFReset bool

	// On the HP48 (SCHIP/SCHIPC) scrolling in lores mode only scrolls half the pixels
	ScaleScroll bool
}

func (q *Quirks) opMem(c *CPU, x uint16) {
	if q.MemLeaveI {
		// Do nothing (Superchip 1.1 behavior)
	} else if q.MemIncIByX {
		// CHIP-48 quirk: increment by X
		c.i += x
	} else {
		// Normal behavior: increment by X + 1
		c.i += x + 1
	}
}
