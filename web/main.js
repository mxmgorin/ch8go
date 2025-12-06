const go = new Go();

async function init() {
  const resp = await fetch("main.wasm");
  const bytes = await resp.arrayBuffer();
  const result = await WebAssembly.instantiate(bytes, go.importObject);

  go.run(result.instance);

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

  window.fillROMs();
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
  const cells = document.querySelectorAll(".key");

  cells.forEach((cell) => {
    const key = cell.textContent.trim().toLowerCase();

    // PRESS
    cell.addEventListener("pointerdown", (e) => {
      e.preventDefault();

      cell.classList.add("pressed"); // <-- add animation class
      window.dispatchEvent(new KeyboardEvent("keydown", { key }));
    });

    // RELEASE
    cell.addEventListener("pointerup", (e) => {
      e.preventDefault();

      cell.classList.remove("pressed"); // <-- remove animation class
      window.dispatchEvent(new KeyboardEvent("keyup", { key }));
    });

    // Safety: pointer leaves key
    cell.addEventListener("pointerleave", (e) => {
      e.preventDefault();

      cell.classList.remove("pressed"); // <-- also remove here
      window.dispatchEvent(new KeyboardEvent("keyup", { key }));
    });
  });
}

// audio
const audioDot = document.getElementById("audio-dot");
let audioCtx = null;
let audioNode = null;
let audioEnabled = false;
let audioFreq = 0.0;
let audioBufSize = 0;
const audioApi = 0;
const audioIconOn = document.getElementById("icon-audio-on");
const audioIconOff = document.getElementById("icon-audio-off");
const audioBtn = document.getElementById("audio");

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
    audioIconOn.style.display = "none";
    audioIconOff.style.display = "inline";
  } else {
    console.log("Audio ON");
    audioNode.connect(audioCtx.destination);
    audioEnabled = true;
    audioIconOn.style.display = "inline";
    audioIconOff.style.display = "none";
  }
}

// scale
const scaleInput = document.getElementById("scaleInput");
const scaleMinus = document.getElementById("scaleMinus");
const scalePlus = document.getElementById("scalePlus");
scaleMinus.onclick = () => {
  let v = parseInt(scaleInput.value);
  if (v > 1) {
    scaleInput.value = v - 1;
    scaleInput.dispatchEvent(new Event("input"));
  }
};
scalePlus.onclick = () => {
  let v = parseInt(scaleInput.value);
  if (v < 15) {
    scaleInput.value = v + 1;
    scaleInput.dispatchEvent(new Event("input"));
  }
};

// settings
const settingsPanel = document.getElementById("settings-panel");
const settingsBtn = document.getElementById("settings-btn");
settingsBtn.onclick = () => {
  if (settingsPanel.style.display == "none") {
    settingsPanel.style.display = "flex";
  } else {
    settingsPanel.style.display = "none";
  }
};

init();
