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

const audioDot = document.getElementById("audio-dot");
let audioCtx = null;
let audioNode = null;
let audioEnabled = false;
let audioFreq = 0.0;
let audioBufSize = 0;
const audioApi = 0;

const iconOn = `
  <svg width="28" height="28" viewBox="0 0 8 8" shape-rendering="crispEdges">
    <!-- body -->
    <rect x="0" y="3" width="2" height="3" fill="black" />
    <rect x="2" y="2" width="1" height="5" fill="black" />
    <!-- inner wave -->
    <rect x="4" y="4" width="1" height="1" fill="black" />
    <!-- center wave -->
    <rect x="5" y="3" width="1" height="1" fill="black" />
    <rect x="5" y="5" width="1" height="1" fill="black" />
    <!-- outer wave -->
    <rect x="6" y="2" width="1" height="1" fill="black" />
    <rect x="6" y="4" width="1" height="1" fill="black" />
    <rect x="6" y="6" width="1" height="1" fill="black" />
  </svg>
`;
const iconOff = `
  <svg width="28" height="28" viewBox="0 0 8 8" shape-rendering="crispEdges">
    <!-- body -->
    <rect x="0" y="3" width="2" height="3" fill="black" />
    <rect x="2" y="2" width="1" height="5" fill="black" />
    <!-- cross 3×3 -->
    <!-- up -->
    <rect x="4" y="3" width="1" height="1" fill="red" />
    <rect x="6" y="3" width="1" height="1" fill="red" />
    <!-- center -->
    <rect x="5" y="4" width="1" height="1" fill="red" />
    <!-- bottom -->
    <rect x="4" y="5" width="1" height="1" fill="red" />
    <rect x="6" y="5" width="1" height="1" fill="red" />
  </svg>
`;
const audioBtn = document.getElementById("audio");
audioBtn.innerHTML = iconOff;
audioBtn.onclick = async () => {
  if (!audioCtx) {
    if (audioApi === 0) {
      await startAudioScriptProcessor();
    } else if (audioApi === 1) {
      await startAudioWorklet();
    }
  }

  toggleAudio();
};

async function startAudioScriptProcessor() {
  audioBufSize = 512;
  window.startAudio(audioBufSize);

  audioCtx = new AudioContext();
  audioFreq = audioCtx.sampleRate;
  console.log("Audio sample rate:", audioFreq);
  await audioCtx.resume();

  audioNode = audioCtx.createScriptProcessor(audioBufSize, 0, 1);
  audioNode.onaudioprocess = (e) => {
    const out = e.outputBuffer.getChannelData(0);
    window.fillAudio(out, audioFreq);
  };
}

async function startAudioWorklet() {
  audioBufSize = 128;
  window.startAudio(audioBufSize);

  audioCtx = new AudioContext();
  audioFreq = audioCtx.sampleRate;
  console.log("Audio sample rate:", audioFreq);

  await audioCtx.audioWorklet.addModule("audio-processor.js");
  audioNode = new AudioWorkletNode(audioCtx, "simple-processor");

  audioNode.port.onmessage = () => {
    const buf = new Float32Array(audioBufSize);
    window.fillAudio(buf, audioFreq);
    audioNode.port.postMessage(buf);
  };
}

function toggleAudio() {
  if (audioEnabled) {
    console.log("Audio OFF");
    audioNode.disconnect();
    audioEnabled = false;
    audioBtn.innerHTML = iconOff;
  } else {
    console.log("Audio ON");
    audioNode.connect(audioCtx.destination);
    audioEnabled = true;
    audioBtn.innerHTML = iconOn;
  }
}

init();
