# <img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/web/icons/favicon.ico" width="22"> ch8go

[![CI](https://github.com/mxmgorin/ch8go/actions/workflows/test.yml/badge.svg)](https://github.com/mxmgorin/ch8go/actions)
[![Go](https://img.shields.io/github/go-mod/go-version/mxmgorin/ch8go)](go.mod)
[![codecov](https://codecov.io/gh/mxmgorin/ch8go/branch/main/graph/badge.svg)](https://codecov.io/gh/mxmgorin/ch8go)
[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://mxmgorin.github.io/ch8go/web/)
[![License: GPL v3](https://img.shields.io/badge/license-GPLv3-blue)](LICENSE)

<img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/super-neat-boy.gif" width="30%"> <img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/alien-inv8sion.gif" width="30%"> <img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/octopeg-big.gif" width="30%">

**`ch8go`** is an accurate CHIP-8, SUPER-CHIP and XO-CHIP emulator written in Go, with broad ROM compatibility and four frontends — the browser (WebAssembly/PWA), SDL2, Ebiten, and an interactive CLI debugger. It passes the full [Timendus](https://github.com/Timendus/chip8-test-suite) and [Octo](https://github.com/JohnEarnest/Octo) test suites, verified in CI.

It began as a project to learn Go and explore the system. *More about CHIP-8 and the design is in the [wiki](https://github.com/mxmgorin/ch8go/wiki).*

🌐 **[Try the Live Demo →](https://mxmgorin.github.io/ch8go/web/)**

## Features

- **Auto configuration** — quirks, tick rate, and color palette are set per-ROM from a built-in [metadata database](https://github.com/chip-8/chip-8-database), so games run correctly without manual tuning.
- **Verified accuracy** — passes the full Timendus and Octo test suites, checked frame-by-frame against golden images in CI ([details](https://github.com/mxmgorin/ch8go/wiki/Testing)).

## Emulation Core

- **CHIP-8**: Implements all 35 standard opcodes, including timers, stack, and registers.
- **SUPER-CHIP**: Implements extended opcodes, high-resolution mode, 16×16 sprites, scrolling, and additional font.
- **XO-CHIP**: Implements extended opcodes, four-plane graphics with 16 colors, and extended audio.
- **Quirks**: Implements all common quirks — shift behavior, jump offsets, VF reset, screen clipping, memory behavior, VBlank waiting, half scrolling ([details](https://github.com/mxmgorin/ch8go/wiki/CHIP%E2%80%908-System#quirks)).

## Frontends

- **WASM** — runs in the browser; installable as a PWA.
- **SDL2** — native desktop, hardware-accelerated.
- **Ebiten** — native desktop, pure Go with no runtime dependencies.
- **CLI** — headless REPL debugger with a disassembler and ASCII renderer.

## Install & Run

Grab a prebuilt binary from the [Releases](https://github.com/mxmgorin/ch8go/releases) page — CLI and Ebiten builds are self-contained; the SDL2 archive bundles its runtime. Prefer no install? Run it in the [browser](https://mxmgorin.github.io/ch8go/web/).

### Build from source

Requires **Go 1.25+**. Run any frontend directly with a ROM:

```bash
go run ./cmd/cli    --rom path/to/game.ch8   # headless REPL/debugger, Go only
go run ./cmd/ebiten --rom path/to/game.ch8   # native GUI, pure Go
go run ./cmd/sdl2   --rom path/to/game.ch8   # native GUI via SDL2
```

- **Ebiten / SDL2** are GUI apps and need the desktop graphics/audio toolchain. On Linux, install the dev headers (see the [Ebitengine install guide](https://ebitengine.org/en/documents/install.html)); **SDL2** also needs its dev package — `libsdl2-dev` (Debian/Ubuntu), `brew install sdl2` (macOS), or `sdl2` (Arch).
- **WASM**: `make wasm` builds `web/main.wasm` and serves the demo at `http://localhost:8000`.

## Controls

```
CHIP-8              Keyboard
1  2  3  C   →      1  2  3  4
4  5  6  D   →      Q  W  E  R
7  8  9  E   →      A  S  D  F
A  0  B  F   →      Z  X  C  V
```

## CLI Usage

<img src="https://raw.githubusercontent.com/mxmgorin/ch8go/main/assets/cli-demo.gif" width="70%">

Run the CLI with:

```bash
go run ./cmd/cli --rom path/to/game.ch8
```

Inside the prompt (`ch8go>`), type `help` for the commands below.

<details>
<summary><b>Commands</b> — full REPL reference</summary>

| Command          | Description                                         |
| ---------------- | --------------------------------------------------- |
| `help`           | Show all supported commands                         |
| `load <file>`    | Load a ROM into memory                              |
| `info`           | Show metadata about a ROM                           |
| `step <n>`       | Execute 1 or N instructions                         |
| `peek <n>`       | Disassemble 1 or N instructions starting from PC    |
| `dis`            | Disassemble the loaded ROM                          |
| `regs`           | Show registers, call stack, and delay timer         |
| `mem <addr> [n]` | Hex-dump N bytes of memory from `addr` (default 64) |
| `draw`           | Render the current display buffer in ASCII          |
| `break <addr>`   | Set a breakpoint at `addr` (hex `0x300` or decimal) |
| `breaks`         | List breakpoints                                    |
| `delete [addr]`  | Remove a breakpoint (or all if omitted)             |
| `watch <t>`      | Watch a memory `addr` or register (`v0`-`vF`); break on change |
| `watches`        | List watchpoints                                    |
| `unwatch [t]`    | Remove a watchpoint (or all if omitted)             |
| `continue [n]`   | Run until a breakpoint/watchpoint (or `n` steps max); alias `c` |
| `keydown <hex>`  | Press a key (`0`-`F`)                               |
| `keyup <hex>`    | Release a key (`0`-`F`)                             |
| `keys`           | List currently pressed keys                         |
| `quit`           | Exit the REPL                                        |

</details>

<details>
<summary><b>Example</b> — disassemble, break, run to the breakpoint, then inspect state</summary>

```text
ch8go> peek 5
0200: 1436  JP  436
0436: A64A  LD  I, 64A
0438: 2378  CALL 378
0378: 6A00  LD  VA, 00
037A: 6B00  LD  VB, 00

ch8go> break 0x37e
Breakpoint set at 037E.

ch8go> continue
Hit breakpoint after 6 steps.
037E: DAB8  DRW VA, VB, 8

ch8go> regs
PC=037E I=064A SP=01 DT=00
V=[0 0 0 0 0 0 0 0 0 0 0 0 8 0 0 0]
Stack=[043A]

ch8go> mem 0x64a 16
064A  61 81 A0 81 A0 81 A0 81 FF 00 00 00 00 00 00 00 |a...............|
```

</details>

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
