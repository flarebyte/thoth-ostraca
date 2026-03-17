# Use Cases

Migrated use-case catalog from docs/designs/use_cases.mts.

## Status

Implementation status used in the use-case catalog.

### Legend

#### Status legend

- `completed`: delivered in the shipped CLI
- `partial`: available, but narrower or less practical than the design intent
- `missing`: described or needed, but not delivered in the shipped CLI
- `exceeds`: delivered beyond the original design expectation

## Use Cases

All modeled use cases as individual notes.

### Catalog

#### Capture stage boundary fixtures

Status: completed.

Dump input/output JSON/NDJSON for reproducible debugging.

#### Create many meta files

Status: partial.

#### Decode shell JSON automatically

Status: missing.

Treat JSON shell output as structured data by default instead of exposing only
raw stdout strings.

#### Diagnose a single stage

Status: completed.

Execute one pipeline stage in isolation with explicit or prepared input;
capture fixtures.

#### Diff meta files at scale

Status: partial.

#### Direct shell field interpolation

Status: missing.

Shell templating should support direct field interpolation beyond a single
`{json}` placeholder.

#### Dry-run save mode

Status: missing.

Save/update workflows should support a clear dry-run mode before writing
sidecars.

#### Expose Git metadata for inputs

Status: partial.

Use go-git to provide tracked/ignored, worktree status, and last commit info
when enabled.

#### Expose os.FileInfo for inputs

Status: partial.

Include size, mode, modTime, isDir for filtering/mapping when enabled.

#### File pipeline: filter, shell, map, persist

Status: missing.

Discover input files, filter them, run shell analysis per file, map the
results, and either emit JSON or persist `.thoth.yaml` sidecars.

#### Filter meta by locator

Status: completed.

Boolean predicate over {locator, meta}.

#### Helpful, well-documented flags

Status: partial.

#### JSON output for CLI/CI/AI

Status: completed.

Machine-oriented default; aggregated JSON; lines optional.

#### Load action config file

Status: completed.

Prefer CUE (.cue) with schema validation.

#### Locators as file path or URL

Status: partial.

#### Lua limits allow ordinary transforms

Status: missing.

Default Lua limits should allow normal transformation logic without immediate
instruction-limit failures.

#### Map meta records

Status: completed.

Transform {locator, meta} -> any.

#### One file per locator

Status: partial.

Minimize merge conflicts.

#### Process in parallel

Status: completed.

Goroutines + channels; bounded pool; default workers = CPU count.

#### Progress reporting for long-running actions

Status: missing.

Long-running actions should expose user-visible progress.

#### Reduce across meta set

Status: completed.

Aggregate stream -> single result.

#### Respect .gitignore by default

Status: partial.

Always on; opt-out via --no-gitignore.

#### Run shell using map output

Status: partial.

Support bash, sh, zsh early.

#### Safe discovery default excludes

Status: missing.

Discovery should exclude `.git`, fixture trees, and similar internal
directories by default.

#### Script filter/map/reduce

Status: partial.

Lua only (v1): small + popular.

#### Write sidecars to a dedicated output directory

Status: missing.

Generated `.thoth.yaml` files should be able to go into a separate output tree
instead of beside source files.

#### Update many meta files

Status: partial.

#### Validate {locator, meta} schema

Status: completed.

Required fields: locator:string, meta:object; error on missing.

#### Validate meta files only

Status: completed.

No transforms or shell; emit validation report.

