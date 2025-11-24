#!/usr/bin/env bash
set -e

echo "Building WASM..."
GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/wasm

cd web
python3 -m http.server 8000
