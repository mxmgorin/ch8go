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
    chip8_loadROM(data);
  };

  const select = document.getElementById("roms");
  select.onchange = async (e) => {
    const path = e.target.value;
    if (path) loadRomFromUrl(path);
  };

  const firstRom = select.options[0].value;
  await loadRomFromUrl(firstRom);
}

async function loadRomFromUrl(url) {
  const resp = await fetch(url);
  const buf = new Uint8Array(await resp.arrayBuffer());
  chip8_loadROM(buf);
}

init();
