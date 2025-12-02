const go = new Go();

async function init() {
  const resp = await fetch("main.wasm");
  const bytes = await resp.arrayBuffer();
  const result = await WebAssembly.instantiate(bytes, go.importObject);

  go.run(result.instance);
  console.log("WASM started");

  document.getElementById("romInput").onchange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const data = new Uint8Array(await file.arrayBuffer());
    chip8_loadROM(data, file.name);
  };

  const input = document.getElementById("romInput");
  const fileName = document.getElementById("fileName");

  input.addEventListener("change", () => {
    fileName.textContent = input.files[0]?.name || "";
  });

  const select = document.getElementById("roms");
  select.onchange = async (e) => {
    const path = e.target.value;
    fileName.textContent = e.target.value;
    if (path) loadRomFromUrl(path);
  };

  const firstRom = select.options[0].value;
  fileName.textContent = firstRom;
  await loadRomFromUrl(firstRom);

  setupKeyboard();
}

async function loadRomFromUrl(url) {
  const resp = await fetch(url);
  const buf = new Uint8Array(await resp.arrayBuffer());
  chip8_loadROM(buf, url);
}

function setupKeyboard() {
  const cells = document.querySelectorAll(".ti-keys .key");

  cells.forEach((cell) => {
    const key = cell.textContent.trim().toLowerCase();

    // PRESS
    cell.addEventListener("pointerdown", (e) => {
      e.preventDefault();
      window.dispatchEvent(new KeyboardEvent("keydown", { key }));
    });

    // RELEASE
    cell.addEventListener("pointerup", (e) => {
      e.preventDefault();
      window.dispatchEvent(new KeyboardEvent("keyup", { key }));
    });

    // Safety: handle pointer leaving the key while still pressed
    cell.addEventListener("pointerleave", (e) => {
      e.preventDefault();
      window.dispatchEvent(new KeyboardEvent("keyup", { key }));
    });
  });
}
const dot = document.getElementById("audio-dot");
let ctx = null;
let node = null;
let on = false;
let sampleRate = 48000.0

document.getElementById("audio").onclick = async () => {
  if (!ctx) {
    ctx = new AudioContext();
    sampleRate = ctx.sampleRate;
    console.log("Audio sample rate:", ctx.sampleRate);
    await ctx.audioWorklet.addModule("audio-processor.js");
    node = new AudioWorkletNode(ctx, "simple-processor");

    // Worklet requests audio → fill buffer via Go/WASM
    node.port.onmessage = () => {
      const buf = new Float32Array(128);
      window.fillAudio(buf, sampleRate);
      node.port.postMessage(buf);
    };
  }

  if (on) {
    node.disconnect();
    on = false;
    dot.classList.remove("on");
    console.log("Audio OFF");
  } else {
    node.connect(ctx.destination);
    on = true;
    dot.classList.add("on");
    console.log("Audio ON");
  }
};

init();
