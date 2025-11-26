[![CI](https://github.com/mxmgorin/ch8go/actions/workflows/test.yml/badge.svg)](https://github.com/mxmgorin/ch8go/actions)

# ch8go

`ch8go` is a CHIP-8 and Super-CHIP interpreter (emulator) written in Go, created as a fun project to practice the language and explore the system. The goal is to implement an accurate CHIP-8 system with good ROM compatibility and support for Super-CHIP and XO-CHIP.

🌐 [Try the Live Demo](https://mxmgorin.github.io/ch8go/web/)

## Features

- **CHIP-8 Support**: Implements all 35 standard opcodes, including timers, stack, and registers.
- **SUPER-CHIP Support**: Implements extended opcodes, high-resolution mode, 16×16 sprites, scrolling, and additional font.
- **Quirks Support**: Implements all common CHIP-8 / SCHIP / XO quirks, including shift behavior, jump offsets, VF reset, screen clipping, memory increment behavior, and VBlank timings.
- **CLI Frontend**: Runs headless with an interactive REPL/debugger, built-in disassembler, and ASCII display renderer.
- **WASM Frontend**: Runs directly in the browser using WebAssembly.
- **SDL2 Frontend**: Runs natively using hardware-accelerated graphics.
- **Integration Tests**: Uses Go’s testing framework and CI to run community-made test ROMs.

## Controls

```
CHIP-8              Keyboard
1  2  3  C   →      1  2  3  4
4  5  6  D   →      Q  W  E  R
7  8  9  E   →      A  S  D  F
A  0  B  F   →      Z  X  C  V
```

## CLI Usage

Run the CLI with:

```bash
go run ./cmd/cli --rom path/to/game.ch8
```

Inside the prompt (`ch8go>`), you can use the following commands:

| Command       | Description                                      |
| ------------- | ------------------------------------------------ |
| `help`        | Show all supported commands                      |
| `load <file>` | Load a ROM into memory                           |
| `step <n>`    | Execute 1 or N instructions                      |
| `peek <n>`    | Disassemble 1 or N instructions starting from PC |
| `regs`        | Show registers                                   |
| `dis`         | Disassemble the loaded ROM                       |
| `draw`        | Render the current display buffer in ASCII       |
| `info`        | Show metadata about a ROM                        |
| `quit`        | Exit the REPL                                    |

**Example session**

```bash
❯ go run ./cmd/cli --rom ./roms/test_opcode.ch8
ch8go CLI. Type 'help' for commands.
ROM loaded (478 bytes).

ch8go> peek 2
0200: 124E  JP  24E
024E: 6801  LD  V8, 01

ch8go> step
0200: 124E  JP  24E

ch8go> regs
PC=03DC I=0202 V=[1 3 7 0 0 42 137 236 44 48 52 26 0 0 0 0]
```

## References

Useful resources for CHIP-8 development:

- [Technical Reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM)
- [Chip-8 on the COSMAC VIP](https://www.laurencescotford.net/2020/07/25/chip-8-on-the-cosmac-vip-index)
- [Opcode Table](https://chip8.gulrak.net)
- [Instruction Set](https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set)
- [Timendus' test ROMS](https://github.com/Timendus/chip8-test-suite)
- [Mastering SuperChip](https://johnearnest.github.io/Octo/docs/SuperChip.html)
- [CHIP-8 Database](https://github.com/chip-8/chip-8-database)
- [CHIP-8 Archive](https://johnearnest.github.io/chip8Archive/)
- [CHIP-8 Research Facility](https://chip-8.github.io/)
- [Awesome CHIP-8](https://github.com/tobiasvl/awesome-chip-8)
