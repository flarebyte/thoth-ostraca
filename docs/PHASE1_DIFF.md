# Phase 1 vs FLOW_DESIGN.md (Walking Skeleton Only)

This document highlights where the Phase 1 implementation (walking skeleton) intentionally differs from, simplifies, or matches the full design described in `docs/designs/FLOW_DESIGN.md`. The focus is strictly on the skeleton pipeline (validate-config → discover → parse/validate YAML → lua-filter → lua-map → shell-exec → lua-postmap → lua-reduce → output), not future flows like create/update/diff.

## Scope Alignment
- Implemented commands: `thoth run`, `thoth diagnose`, `thoth version` (dev).
- Implemented pipeline stages: `validate-config`, `discover-meta-files`, `parse-validate-yaml`, `lua-filter`, `lua-map`, `shell-exec`, `lua-postmap`, `lua-reduce`.
- Output modes: aggregated JSON envelope; NDJSON lines.
- Error modes: `errors.mode` (fail-fast|keep-going) and `errors.embedErrors`.
- Parallelism: bounded worker pool on per-record stages with deterministic reassembly.
- E2E tests: golden/NDJSON outputs; determinism checks.

## Key Differences vs. FLOW_DESIGN.md

### Config Schema and Versioning
- Full spec: CUE schema with `configVersion: "1"` and richer sections (validation, lua limits, locatorPolicy, etc.).
- Phase 1: Minimal CUE extraction only for:
  - `configVersion` (string), `action` (string)
  - Optional: `discovery.root`, `discovery.noGitignore`
  - Optional: `filter.inline`, `map.inline`, `postMap.inline`, `reduce.inline`
  - Optional: `shell.enabled`, `shell.program`, `shell.argsTemplate`, `shell.timeoutMs`
  - Optional: `output.lines`
  - Optional: `errors.mode`, `errors.embedErrors`
  - Optional: `workers`
- Deviation: No strict schema file or version policy; examples and tests use `configVersion: "v0"` rather than `"1"`.
- Deviation: No `validation.*`, `lua.*` limits, or `locatorPolicy.*` handling.

### Discovery
- Full spec: Walk + gitignore filter; apply `.gitignore` even when not a Git repo; no symlink following by default.
- Phase 1: Walk via `filepath.WalkDir`; parses and applies `.gitignore` patterns from all ancestor dirs using `go-git`'s ignore matcher; sorts locators; finds only `*.thoth.yaml` files.
- Not included: followSymlinks, OS/Git enrichment.

### YAML Parse/Validate
- Full spec: Validate required fields; options for unknown keys and policies.
- Phase 1: Strictly require top-level mapping and required fields `locator` (string) and `meta` (object). Errors are short and consistent. No additional validation knobs.

### Lua (filter/map/reduce/postmap)
- Full spec: Configurable limits (timeout/instruction/memory), deterministic random, helper namespace, etc.
- Phase 1: Minimal sandbox: open `base`, `string`, `table`, `math` only; fixed 200ms timeout per call; no instruction/memory limits; no deterministic RNG or extra helpers. Filter/map accept expressions without explicit `return`.

### Shell Execution
- Full spec: `workingDir`, `env`, `killProcessGroup`, `termGraceMs`, capture toggles, size limits, strict templating.
- Phase 1: Minimal: `enabled`, `program`, `argsTemplate` with `{json}` substitution, `timeoutMs`; capture `stdout`, `stderr`, and `exitCode`. No PG kill, env/dir, or capture limits.

### Reduce
- Full spec: Deterministic over sorted items; any JSON value; ignore `--lines` when reduce present.
- Phase 1: Deterministic; default reduce = count of records when no inline provided; adhere to `--lines` behavior when no reduce is used.

### Ordering and Determinism
- Full spec: Deterministic ordering for aggregated outputs; lines may be nondeterministic due to parallelism.
- Phase 1: Deterministic ordering throughout:
  - `discover-meta-files`: sorts locators.
  - `parse-validate-yaml`: collects and sorts by `locator` before emitting records.
  - Per-record parallel stages: use index-aware worker-pool and reassemble in stable order. NDJSON lines are emitted in record order (still deterministic in Phase 1).
- Errors: Phase 1 additionally sorts `envelope.errors` by `(stage, locator, message)` before output for reproducibility.

### Error Handling
- Full spec: Keep-going continues and aggregates errors; fail-fast aborts; optional embedding of per-record errors.
- Phase 1: Matches intent with minimal shape:
  - `errors.mode`: `keep-going` vs `fail-fast`.
  - `errors.embedErrors`: when false, per-record embedded errors are stripped in final output (envelope still lists errors).
  - Exit code rules for keep-going: non-zero if no successful records.

### Diagnose
- Full spec: Stage selection, input resolution, dump input/output, header, etc.
- Phase 1: Implements `--stage`, `--in`, `--dump-in`, `--dump-out`, `--config`, and discovery flags when preparing input.

### Parallelism (Workers)
- Full spec: Default workers = `runtime.NumCPU()`; configurable via config.
- Phase 1: Bounded worker pool applied to per-record stages only (parse-validate-yaml, lua-filter, lua-map, shell-exec, lua-postmap). `workers` honored when present; default used internally but not exposed unless set explicitly in config.

## Explicitly Deferred (Out of Phase 1)
- Flows: `create`, `update`, `diff`, `validate-only` (standalone) — not implemented.
- Enrichment: OS file info, Git metadata exposure for inputs — not implemented.
- Advanced config: validation policy, locator policy, extended Lua limits, pretty output and file destinations.
- Shell: working dir, env, process group control, advanced capture limits.

## Summary of Shifts
- Versioning and schema are looser in Phase 1 (accept "v0"; no hard CUE schema file).
- Determinism is strengthened (sorted `envelope.errors` and deterministic NDJSON lines), which is stricter than the spec’s allowance for nondeterministic line order.
- Lua and shell are minimal, omitting most optional controls from the full spec.
- Workers are supported and deterministic reassembly is enforced; the `workers` value is only surfaced in output when explicitly configured.

These deviations keep the walking skeleton small, testable, and reproducible, while leaving clear extension points to align with the complete FLOW_DESIGN in later phases.
