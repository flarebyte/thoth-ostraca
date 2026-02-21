# Contributing to thoth

Thanks for contributing. This project prioritizes deterministic behavior, stable contracts, and small, reviewable changes.

## Getting Started
- Prerequisites:
  - Go (use the version from `go.mod`)
  - Bun/Node only if you run E2E TypeScript tests
- Build local binary:

```bash
go build -o .e2e-bin/thoth ./cmd/thoth
```

- Run core tests:

```bash
go test ./...
```

- Optional full checks:

```bash
make test
make test-race
make bench
```

## Development Workflow
- Keep PRs small and focused (one behavior change per PR when possible).
- Prefer adding tests first or with the change.
- Avoid unrelated refactors in behavior PRs.
- When changing output behavior, update fixtures/goldens intentionally and explain why.
- Never rely on map iteration order, current time, random values, or worker scheduling for output fields.
- If a change affects user-facing contract, add/adjust contract snapshot tests.

## Testing Commands
- Unit/integration tests:

```bash
go test ./...
```

- Race detector:

```bash
make test-race
```

- Benchmarks:

```bash
make bench
```

- Optional perf smoke:

```bash
make perf-smoke
```

## Golden and Fixture Policy
- Repos/fixtures:
  - Add scenario repos under `testdata/repos/<name>/`.
  - Keep fixtures minimal and purpose-specific.
  - Prefer deterministic file names and contents.
- Config fixtures:
  - Add or update configs under `testdata/configs/`.
- Goldens:
  - Store run outputs under `testdata/run/*.golden.json` (or `.golden.ndjson`).
  - Update only when behavior change is intentional.
  - Validate by rerunning tests, not by manual editing alone.
- Determinism requirement:
  - Output must be byte-identical across reruns.
  - Output must be byte-identical for equivalent runs with different workers (e.g. `workers=1` vs `workers=8`) unless explicitly documented otherwise.

## Code Style Expectations
- Keep stages small and single-purpose.
- Prefer explicit, readable helpers over large monolithic functions.
- Errors must be short, single-line, and deterministic.
- Include locator/path in errors when available.
- Do not introduce nondeterministic fields into envelope/record output.
- Keep machine-oriented defaults stable (stdout JSON contract first).

## PR Checklist
- [ ] Change scope is small and focused.
- [ ] Added/updated tests for behavior change.
- [ ] `go test ./...` passes.
- [ ] If relevant: `make test-race` passes.
- [ ] If relevant: `make bench` compiles/runs.
- [ ] Goldens updated intentionally and reviewed.
- [ ] Output remains deterministic (rerun + workers comparison).
- [ ] Error messages remain short, stable, and deterministic.

## Notes
- Use `.e2e-bin/thoth` for tests and local verification.
- Do not commit built binaries.
