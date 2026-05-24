## Makefile: thin wrappers for local/dev/release commands.
## Keep target behavior explicit and deterministic.
## Build artifacts are produced under ./build by build-go.ts.
## Release publishing is handled by release-go.ts.

.PHONY: lint format test test-race gen docs-gen build build-dev e2e release clean help bench perf-smoke contract-snapshots release-check ghf-sync ghf-check ghf-test ghf-cov ghf-format ghf-release

BIOME := npx @biomejs/biome
BUN := bun
GO := go
GOLINT := golangci-lint
FLYB := flyb
GHF := gh flarebyte

lint:
	$(BIOME) check
	$(GO) vet ./...
	$(GOLINT) run

format:
	gofmt -w .
	$(BIOME) format --write .
	$(BIOME) check --unsafe --write

test: gen ghf-test

test-race: gen
	$(GO) test -race ./...

bench:
	$(GO) test -bench=. -run=^$$ ./...

perf-smoke:
	$(GO) test -run TestPerfSmoke_ ./...

contract-snapshots:
	$(GO) test -run TestContract_ ./internal/stage

gen:
	$(MAKE) docs-gen

docs-gen:
	$(FLYB) validate --config docs/design-meta
	$(FLYB) generate markdown --config docs/design-meta

build:
	$(GHF) build

build-dev:
	mkdir -p .e2e-bin
	GOCACHE=$(PWD)/.gocache GOMODCACHE=$(PWD)/.gomodcache CGO_ENABLED=0 $(GO) build -o .e2e-bin/thoth ./cmd/thoth

e2e:
	cd script/e2e && $(BUN) test

release: release-check
	$(BUN) run release-go.ts

ghf-sync:
	$(GHF) sync

ghf-check:
	$(GHF) check

ghf-test:
	$(GHF) test

ghf-cov:
	$(GHF) cov

ghf-format:
	$(GHF) format

ghf-release:
	$(GHF) release

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

review: format test ghf-cov e2e lint

thoth-meta: thoth-meta-go thoth-meta-go-test thoth-meta-ts-e2e

thoth-meta-go:
	./.e2e-bin/thoth run --config ./pipeline-go-maat.thoth.cue

thoth-meta-go-test:
	./.e2e-bin/thoth run --config ./pipeline-go-test-maat.thoth.cue

thoth-meta-ts-e2e:
	./.e2e-bin/thoth run --config ./pipeline-ts-e2e-maat.thoth.cue 

thoth-lint-go:
	./.e2e-bin/thoth run --config ./pipeline-go-function-thresholds.thoth.cue
	cat temp/pipeline-go-function-thresholds.json | jq '.meta.reduced.worstOffenders'

thoth-meta-merge:
	./.e2e-bin/thoth run --config ./pipeline-thoth-meta-aggregate.thoth.cue
	
help:
	@printf "Targets:\n"
	@printf "  (requires: go, bun, golangci-lint, biome)\n"
	@printf "  lint               Run linters (Biome + go vet + golangci-lint).\n"
	@printf "  format             Apply formatting (gofmt + Biome).\n"
	@printf "  test               Run Go unit tests.\n"
	@printf "  test-race          Run Go tests with race detector.\n"
	@printf "  bench              Run Go benchmarks.\n"
	@printf "  perf-smoke         Run performance smoke tests.\n"
	@printf "  contract-snapshots Run contract snapshot tests.\n"
	@printf "  release-check      Run lint + tests + contract snapshots.\n"
	@printf "  gen                Generate repo artifacts.\n"
	@printf "  docs-gen           Generate design docs from flyb config.\n"
	@printf "  build              Run gh-flarebyte build workflow.\n"
	@printf "  build-dev          Build local dev binary into .e2e-bin/.\n"
	@printf "  e2e                Run Bun-powered end-to-end tests.\n"
	@printf "  release            Run release checks, build artifacts, publish GitHub release.\n"
	@printf "  clean              Remove build artifacts.\n"
	@printf "  complexity         Show top file complexity (Go/TS).\n"
	@printf "  sec                Run security scan (semgrep).\n"
	@printf "  dup                Run duplication scans (go/typescript).\n"
	@printf "  review             Run format, tests, coverage, e2e, and lint.\n"
	@printf "  ghf-sync           Sync repository governance/config via gh-flarebyte.\n"
	@printf "  ghf-check          Run gh-flarebyte checks.\n"
	@printf "  ghf-test           Run gh-flarebyte test workflow.\n"
	@printf "  ghf-cov            Run gh-flarebyte coverage workflow.\n"
	@printf "  ghf-format         Run gh-flarebyte formatting workflow.\n"
	@printf "  ghf-release        Run gh-flarebyte release workflow.\n"
