
.PHONY: test bench cover clean lint

# Run only tests (no benchmarks)
test:
	go test ./pkg/host -bench=^$

# Run tests with coverage across all pkg packages.
# Uses binary coverage + covdata so cross-package profiles (e.g. host tests
# exercising chip8) are merged correctly instead of double-counted.
cover:
	rm -rf .coverdir && mkdir -p .coverdir
	go test -cover -coverpkg=./pkg/... ./pkg/... -args -test.gocoverdir=$(CURDIR)/.coverdir
	go tool covdata textfmt -i=.coverdir -o=coverage.out
	go tool covdata percent -i=.coverdir
	rm -rf .coverdir

# Run only benchmarks (no tests)
bench:
	go test ./pkg/host -run=^$$ -bench=. -count=1

# Regenerate PNG output files for tests
test-update:
	go test ./pkg/host -run=. -bench=^$ -- -update-golden

# Remove generated PNG outputs
test-clean:
	rm -rf testdata/golden/*

# static analysis
lint:
	go vet ./...

wasm:
	@echo "Building WASM..."
	GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/wasm

	@echo "Starting local server on http://localhost:8000"
	cd web && python3 -m http.server 8000
