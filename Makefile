
.PHONY: test bench clean lint

# Run only tests (no benchmarks)
test:
	go test ./app -bench=^$

# Run only benchmarks (no tests)
bench:
	go test ./app -run=^$$ -bench=. -count=1

# Regenerate PNG output files for tests
test-output:
	go test ./app -run=. -bench=^$ -- -output-png

# Remove generated PNG outputs
test-clean:
	rm -rf app/testdata/output/*

# static analysis
lint:
	go vet ./...

wasm:
	@echo "Building WASM..."
	GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/wasm

	@echo "Starting local server on http://localhost:8000"
	cd web && python3 -m http.server 8000
