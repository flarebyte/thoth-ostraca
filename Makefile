## Makefile: Thin, explicit wrappers for tools
## - One responsibility per target
## - No dynamic variables or shell logic
## - Real logic lives in scripts (TypeScript/Bun, bash, Go)

.PHONY: lint format test gen build release clean help

BIOME := npx @biomejs/biome
BUN := bun
GO := go

lint:
	$(BIOME) check
	$(GO) vet ./...

format:
	gofmt -w .
	$(BIOME) format --write .
	$(BIOME) check --write

test: gen
	$(GO) test ./...

gen:
	true

build:
	$(BUN) run build-go.mts

release: build
	@printf "Artifacts in ./build (checksums.txt included)\n"

clean:
	rm -rf build

help:
	@printf "Targets:\n"
	@printf "  lint     Run linters (Biome + go vet).\n"
	@printf "  format   Apply formatting (gofmt + Biome).\n"
	@printf "  test     Run Go tests.\n"
	@printf "  gen      Generate artifacts (no-op placeholder).\n"
	@printf "  build    Build Go binaries via Bun script.\n"
	@printf "  release  Prepare release artifacts (depends on build).\n"
	@printf "  clean    Remove build artifacts.\n"
