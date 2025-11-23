# ch8go

A CHIP-8 emulator written in Go.
The goal of this project is to have fun and practice Go while exploring CHIP-8.

🌐 [Try the Live Demo here](https://mxmgorin.github.io/ch8go/web/)

## Features

- **Full CHIP-8 CPU Emulation**: Implements all standard opcodes, including timers, stack, and registers.
- **Built-in REPL / CLI Debugger**: Step through instructions, inspect registers, view display buffer, load ROMs, and run commands interactively.
- **Disassembler**: Convert CHIP-8 ROMs to readable assembly.
- **ASCII Display Renderer**: Render the CHIP-8 display directly in the terminal.
- **SDL2 Frontend**: Hardware-accelerated graphics window.
- **WebAssembly Frontend (WASM)**: Run the emulator in the browser using WASM. Includes HTML/JS bindings for display output, keyboard input, and ROM loading.

## CLI Usage

The project includes an interactive CLI / REPL for debugging and inspecting CHIP-8 programs.
Start it with:

```bash
go run ./cmd/cli --rom path/to/game.ch8
```

Once inside the prompt (chip8>), you can type help at any time to see all available commands.
| Command | Description |
| ------------- | ----------------------------------------------------------------- |
| `help` | Show a list of all supported commands. |
| `load <file>` | Load a CHIP-8 ROM into memory. |
| `step` | Execute a single instruction. |
| `run <steps>` | Execute multiple instructions (default: 10). |
| `regs` | Print registers. |
| `disasm <n>` | Disassemble the next _n_ instructions starting at the current PC. |
| `draw` | Render the current display buffer in ASCII. |
| `quit` | Exit the REPL. |

**Example session**

```bash
❯ go run ./cmd/cli --rom ./roms/test_opcode.ch8
CHIP-8 CLI. Type 'help' for commands.
ROM loaded (478 bytes).

chip8> disasm 3
0200: 124E  JP  24E
024E: 6801  LD  V8, 01
0250: 6905  LD  V9, 05

chip8> step
0200: 124E  JP  24E

chip8> run 500
Executed 500 steps.

chip8> regs
PC=03DC I=0202 V=[1 3 7 0 0 42 137 236 44 48 52 26 0 0 0 0]
```

## References

Useful resources for CHIP-8 development:

- [Technical Reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM)
- [Instruction Set](https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set)
- [Awesome CHIP-8](https://github.com/tobiasvl/awesome-chip-8)
- [CHIP-8 Archive](https://johnearnest.github.io/chip8Archive/)
