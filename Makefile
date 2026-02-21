## Makefile: thin wrappers for local/dev/release commands.
## Keep target behavior explicit and deterministic.
## Release artifacts are produced under ./build by build-go.ts.

.PHONY: lint format test test-race gen build build-dev e2e release clean help bench perf-smoke contract-snapshots release-check

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
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

test-race: gen
	$(GO) test -race ./...

bench:
	$(GO) test -bench=. -run=^$$ ./...

perf-smoke:
	$(GO) test -run TestPerfSmoke_ ./...

contract-snapshots:
	$(GO) test -run TestContract_ ./internal/stage

gen:
	true

build:
	$(BUN) run build-go.ts

build-dev:
	mkdir -p .e2e-bin
	GOCACHE=$(PWD)/.gocache GOMODCACHE=$(PWD)/.gomodcache CGO_ENABLED=0 $(GO) build -o .e2e-bin/thoth ./cmd/thoth

e2e:
	cd script/e2e && $(BUN) test

release: build
	@printf "Artifacts in ./build (checksums.txt included)\n"

release-check: lint test contract-snapshots

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
	@printf "  (requires: go, bun, golangci-lint, biome)\n"
	@printf "  lint               Run linters (Biome + go vet + golangci-lint).\n"
	@printf "  format             Apply formatting (gofmt + Biome).\n"
	@printf "  test               Run Go tests + coverage summary.\n"
	@printf "  test-race          Run Go tests with race detector.\n"
	@printf "  bench              Run Go benchmarks.\n"
	@printf "  perf-smoke         Run performance smoke tests.\n"
	@printf "  contract-snapshots Run contract snapshot tests.\n"
	@printf "  release-check      Run lint + tests + contract snapshots.\n"
	@printf "  gen                Generate artifacts (no-op placeholder).\n"
	@printf "  build              Build release binaries into ./build.\n"
	@printf "  build-dev          Build local dev binary into .e2e-bin/.\n"
	@printf "  e2e                Run Bun-powered end-to-end tests.\n"
	@printf "  release            Build release artifacts and checksums.\n"
	@printf "  clean              Remove build artifacts.\n"
	@printf "  complexity         Show top file complexity (Go/TS).\n"
	@printf "  sec                Run security scan (semgrep).\n"
	@printf "  dup                Run duplication scans (go/typescript).\n"
