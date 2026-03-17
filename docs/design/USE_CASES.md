# Use Cases

Migrated use-case catalog from docs/designs/use_cases.mts.

## Status

Implementation status used in the use-case catalog.

### Legend

#### Status legend

- `completed` label: delivered in the shipped CLI
- `partial` label: available, but narrower or less practical than the design intent
- `missing` label: described or needed, but not delivered in the shipped CLI
- `exceeds` label: delivered beyond the original design expectation

## Use Cases

All modeled use cases as individual notes.

### Catalog

#### Capture stage boundary fixtures

Dump input/output JSON/NDJSON for reproducible debugging.

#### Create many meta files

#### Decode shell JSON automatically

Treat JSON shell output as structured data by default instead of exposing only raw stdout strings.

#### Diagnose a single stage

Execute one pipeline stage in isolation with explicit or prepared input; capture fixtures.

#### Diff meta files at scale

#### Direct shell field interpolation

Shell templating should support direct field interpolation beyond a single `{json}` placeholder.

#### Dry-run save mode

Save/update workflows should support a clear dry-run mode before writing sidecars.

#### Expose Git metadata for inputs

Use go-git to provide tracked/ignored, worktree status, and last commit info when enabled.

#### Expose os.FileInfo for inputs

Include size, mode, modTime, isDir for filtering/mapping when enabled.

#### File pipeline: filter, shell, map, persist

Discover input files, filter them, run shell analysis per file, map the results, and either emit JSON or persist `.thoth.yaml` sidecars.

#### Filter meta by locator

Boolean predicate over {locator, meta}.

#### Helpful, well-documented flags

#### JSON output for CLI/CI/AI

Machine-oriented default; aggregated JSON; lines optional.

#### Load action config file

Prefer CUE (.cue) with schema validation.

#### Locators as file path or URL

#### Lua limits allow ordinary transforms

Default Lua limits should allow normal transformation logic without immediate instruction-limit failures.

#### Map meta records

Transform {locator, meta} -> any.

#### One file per locator

Minimize merge conflicts.

#### Process in parallel

Goroutines + channels; bounded pool; default workers = CPU count.

#### Progress reporting for long-running actions

Long-running actions should expose user-visible progress.

#### Reduce across meta set

Aggregate stream -> single result.

#### Respect .gitignore by default

Always on; opt-out via --no-gitignore.

#### Run shell using map output

Support bash, sh, zsh early.

#### Safe discovery default excludes

Discovery should exclude `.git`, fixture trees, and similar internal directories by default.

#### Script filter/map/reduce

Lua only (v1): small + popular.

#### Write sidecars to a dedicated output directory

Generated `.thoth.yaml` files should be able to go into a separate output tree instead of beside source files.

#### Update many meta files

#### Validate {locator, meta} schema

Required fields: locator:string, meta:object; error on missing.

#### Validate meta files only

No transforms or shell; emit validation report.

