# <img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/web/icons/favicon.ico" width="22"> ch8go

[![CI](https://github.com/mxmgorin/ch8go/actions/workflows/test.yml/badge.svg)](https://github.com/mxmgorin/ch8go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/mxmgorin/ch8go)](https://goreportcard.com/report/github.com/mxmgorin/ch8go)

<img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/super-neat-boy.gif" width="30%"> <img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/alien-inv8sion.gif" width="30%"> <img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/octopeg-big.gif" width="30%">

`ch8go` is a CHIP-8, SUPER-CHIP and XO-CHIP virtual machine (emulator) written in Go, created as a fun project to practice the language and explore the system. The goal is to implement an accurate system with broad ROM compatibility.

*More about CHIP-8 and design notes are available in the [wiki](https://github.com/mxmgorin/ch8go/wiki).*

🌐 [Try the Live Demo](https://mxmgorin.github.io/ch8go/web/)

## Features

- **CHIP-8 Support**: Implements all 35 standard opcodes, including timers, stack, and registers.
- **SUPER-CHIP Support**: Implements extended opcodes, high-resolution mode, 16×16 sprites, scrolling, and additional font.
- **XO-CHIP Support**: Implements extended opcodes, four-plane graphics with 16 colors, and extended audio features.
- **Quirks Support**: Implements all common CHIP-8, SCHIP, XO-CHIP quirks — shift behavior, jump offsets, VF reset, screen clipping, memory increment behavior, VBlank waiting, half scrolling ([details](https://github.com/mxmgorin/ch8go/wiki/CHIP%E2%80%908-System#instruction-quirks)).
- **Ease of Use**: No manual setup required — quirks, tick rate, and color palette are automatically configured for each ROM using a built-in [metadata database](https://github.com/chip-8/chip-8-database)
- **WASM Frontend**: Runs directly in the browser using WebAssembly and is installable as PWA.
- **CLI Frontend**: Runs headless with an interactive REPL/debugger, built-in disassembler, and ASCII display renderer.
- **SDL2 Frontend**: Runs natively using hardware-accelerated graphics.
- **Ebiten Frontend**: Runs natively using pure-Go graphical library with no external runtime dependencies.
- **High accuracy**: Passes all Timendus and Octo test ROMs, executed via `go test` and validated using golden files and continuous integration ([details](https://github.com/mxmgorin/ch8go/wiki/Testing)).

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
- [CHIP-8 Research Facility](https://chip-8.github.io/)
- [Opcode Table](https://chip8.gulrak.net)
- [Instruction Set](https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set)
- [CHIP-8 on the COSMAC VIP](https://www.laurencescotford.net/2020/07/25/chip-8-on-the-cosmac-vip-index)
- [Mastering SUPER-CHIP](https://johnearnest.github.io/Octo/docs/SuperChip.html)
- [XO-CHIP Specification](https://github.com/JohnEarnest/Octo/blob/gh-pages/docs/XO-ChipSpecification.md)
- [Timendus' test ROMS](https://github.com/Timendus/chip8-test-suite)
- [CHIP-8 Database](https://github.com/chip-8/chip-8-database)
- [CHIP-8 Archive](https://johnearnest.github.io/chip8Archive/)
- [Awesome CHIP-8](https://github.com/tobiasvl/awesome-chip-8)

## Acknowledgements

This project is made possible by the work and documentation of the community:
- The original CHIP-8, SUPER-CHIP, and XO-CHIP specifications and their authors, as listed in References section
- Game authors, including participants of [Octo Jam](https://johnearnest.github.io/chip8Archive/), for making their work publicly available
- Test ROM authors, [Timendus](https://github.com/Timendus/chip8-test-suite) and the [Octo](https://github.com/JohnEarnest/Octo) project
- The Go, WebAssembly, Ebiten, and SDL ecosystems for providing robust tooling and libraries 
