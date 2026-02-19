## Makefile: Thin, explicit wrappers for tools
## - One responsibility per target
## - No dynamic variables or shell logic
## - Real logic lives in scripts (TypeScript/Bun, bash, Go)

.PHONY: lint format test test-race gen build build-dev e2e release clean help

BIOME := npx @biomejs/biome
BUN := bun
GO := go
GOLINT := golangci-lint

lint:
	$(BIOME) check
	$(GO) vet ./...
	$(GOLINT) run

format:
	gofmt -w .
	$(BIOME) format --write .
	$(BIOME) check --unsafe --write

test: gen
	$(GO) test ./...

test-race: gen
	$(GO) test -race ./...

gen:
	true

build:
	$(BUN) run build-go.mts

build-dev:
	mkdir -p .e2e-bin
	GOCACHE=$(PWD)/.gocache GOMODCACHE=$(PWD)/.gomodcache CGO_ENABLED=0 $(GO) build -o .e2e-bin/thoth ./cmd/thoth

e2e:
	cd script/e2e && $(BUN) test

release: build
	@printf "Artifacts in ./build (checksums.txt included)\n"

clean:
	rm -rf build

complexity:
	scc --sort complexity --by-file -i go . | head -n 15
	scc --sort complexity --by-file -i ts . | head -n 15

sec:
	semgrep scan --config auto
dup:
	npx jscpd --format go --min-lines 10 --gitignore .
	npx jscpd --format typescript --min-lines 15 --gitignore .

help:
	@printf "Targets:\n"
	@printf "  lint     Run linters (Biome + go vet).\n"
	@printf "  format   Apply formatting (gofmt + Biome).\n"
	@printf "  test     Run Go tests.\n"
	@printf "  test-race Run Go tests with race detector.\n"
	@printf "  gen      Generate artifacts (no-op placeholder).\n"
	@printf "  build    Build Go binaries via Bun script.\n"
	@printf "  build-dev Build dev binary into .e2e-bin/.\n"
	@printf "  e2e      Run Bun-powered end-to-end tests.\n"
	@printf "  release  Prepare release artifacts (depends on build).\n"
	@printf "  clean    Remove build artifacts.\n"
