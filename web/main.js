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

  setupKeyboard();
}

async function loadRomFromUrl(url) {
  const resp = await fetch(url);
  const buf = new Uint8Array(await resp.arrayBuffer());
  chip8_loadROM(buf);
}


function setupKeyboard() {
    const cells = document.querySelectorAll("#keyboard td, #keyboard th");

    cells.forEach(cell => {
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


init();
