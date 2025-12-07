const go = new Go();
const loader = document.getElementById("loader");

async function init() {
  const resp = await fetch("main.wasm");
  const bytes = await resp.arrayBuffer();
  const result = await WebAssembly.instantiate(bytes, go.importObject);

  go.run(result.instance);

  await setupROMS();
  setupKeyboard();
  setupButtons();
  loader.style.display = "none";
}

async function setupROMS() {
  window.fillROMs();
  const roms = document.getElementById("roms");
  const romInput = document.getElementById("romInput");
  const fileName = document.getElementById("fileName");

  romInput.addEventListener("change", async (e) => {
    const file = e.target.files[0];
    if (!file) {
      fileName.textContent = "";
      return;
    }

    fileName.textContent = file.name;
    const data = new Uint8Array(await file.arrayBuffer());
    chip8_loadROM(data, file.name);
  });

  roms.addEventListener("change", async (e) => {
    // save rom to url
    const index = roms.selectedIndex;
    const params = new URLSearchParams(window.location.search);
    params.set("rom", index);
    history.replaceState({}, "", "?" + params.toString());
    // load
    const url = e.target.value;
    fileName.textContent = e.target.value;
    if (url) loadRomFromUrl(url);
  });

  const romParam = getURLParam("rom");
  var rom = roms.options[0].value;
  if (romParam) {
    const option = roms.options[romParam];
    if (option) {
      rom = option.value;
    }
  }

  roms.value = rom;
  fileName.textContent = rom;
  await loadRomFromUrl(rom);
}

function getURLParam(name) {
  const params = new URLSearchParams(window.location.search);
  return params.get(name);
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

function setupButtons() {
  const buttons = document.querySelectorAll(".pressable");

  buttons.forEach((button) => {
    button.addEventListener("pointerdown", (e) => {
      button.classList.add("pressed"); // <-- add animation class
    });

    button.addEventListener("pointerup", (e) => {
      button.classList.remove("pressed"); // <-- remove animation class
    });

    button.addEventListener("pointerleave", (e) => {
      button.classList.remove("pressed"); // <-- also remove here
    });
  });
}

// audio
const audioDot = document.getElementById("audio-dot");
const audioApi = 0;
var audioCtx = null;
var audioEnabled = false;
var audioNode = null;
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
  const audioBufSize = 512;
  window.startAudio(audioBufSize);

  audioCtx = new AudioContext();
  await audioCtx.resume();
  const audioFreq = audioCtx.sampleRate;

  console.log("Audio sample rate:", audioFreq);

  audioNode = audioCtx.createScriptProcessor(audioBufSize, 0, 1);
  audioNode.onaudioprocess = (e) => {
    const out = e.outputBuffer.getChannelData(0);
    window.fillAudio(out, audioFreq);
  };
}

async function startAudioWorklet() {
  const audioBufSize = 128;
  window.startAudio(audioBufSize);
  audioCtx = new AudioContext();
  const audioFreq = audioCtx.sampleRate;

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
    settingsPanel.style.display = null;
  } else {
    settingsPanel.style.display = "none";
  }
};

init();
