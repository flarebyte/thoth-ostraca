# thoth-ostraca

Thoth OSTRACA is a fast, scriptable CLI that discovers, validates, transforms, and aggregates metadata files in a repository. It reads a small CUE configuration and runs a deterministic pipeline over your project’s “ostraca” (metadata shards), producing clean JSON output for downstream tools.

## Overview

The `thoth` CLI processes files that end with `.thoth.yaml` using a fixed pipeline you control via a CUE config (`--config path/to/config.cue`). The pipeline can:

- Discover candidate files (gitignore-aware by default)
- Parse and validate each YAML record (expects `{ locator, meta }`)
- Optionally filter and map records with Lua
- Optionally run a shell command per record (with JSON templating)
- Optionally post-process with Lua and reduce to a final value
- Render either a single JSON envelope or NDJSON lines

The CLI favors determinism and clear error reporting. You choose whether to stop at the first error or keep going and collect errors alongside successful records.

## Features

- Git-aware discovery
  - Walks from a configured root; respects `.gitignore` by default
  - Optional `noGitignore` in config to include ignored files

- YAML parsing + minimal validation
  - Each file must contain a mapping with `locator: string` and `meta: object`
  - Clear, succinct error messages on invalid input

- Lua-powered transforms (optional)
  - `filter`: keep or drop records based on an expression or snippet
  - `map`: transform records into new objects
  - `postMap`: enrich mapped records (e.g., include shell results)
  - `reduce`: fold all records into a single value
  - Snippets can be concise expressions; `return` is added automatically when omitted

- Shell integration (optional)
  - Run a program per record with a timeout
  - Simple argument templating: `{json}` is replaced by the record’s mapped JSON
  - Captures `stdout`, `stderr`, and exit codes

- Error handling modes
  - `keep-going`: continue processing and report errors in the envelope
  - `embedErrors`: when true, include per-record errors inside each record

- Output formats
  - Full JSON envelope (records + meta + errors)
  - NDJSON lines (`output.lines: true`) for easy streaming/grep

- Concurrency with determinism
  - Optional `workers` to parallelize stages
  - Output ordering remains deterministic to match golden results

- Diagnostics
  - `thoth diagnose --stage <name>` runs a single stage
  - Optional `--dump-in` / `--dump-out` to capture inputs/outputs

## Quick Glimpse

- Run the full pipeline:
  - `thoth run --config path/to/config.cue`

- Diagnose a single stage (example):
  - `thoth diagnose --stage validate-config --config path/to/config.cue`

Outputs are written to `stdout`. On failure, a short, actionable message is written to `stderr`.

## Notes

- Configuration is provided as a CUE file (`.cue`).
- File discovery targets `*.thoth.yaml` files.
- The CLI is designed to be predictable: default behavior favors clarity and reproducibility.
