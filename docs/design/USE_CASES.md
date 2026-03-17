# Use Cases

Migrated use-case catalog from docs/designs/use_cases.mts.

## Use Cases

All modeled use cases as individual notes.

### Catalog

#### Capture stage boundary fixtures

Dump input/output JSON/NDJSON for reproducible debugging.

#### Create many meta files

#### Diagnose a single stage

Execute one pipeline stage in isolation with explicit or prepared input; capture fixtures.

#### Diff meta files at scale

#### Expose Git metadata for inputs

Use go-git to provide tracked/ignored, worktree status, and last commit info when enabled.

#### Expose os.FileInfo for inputs

Include size, mode, modTime, isDir for filtering/mapping when enabled.

#### Filter meta by locator

Boolean predicate over {locator, meta}.

#### Helpful, well-documented flags

#### JSON output for CLI/CI/AI

Machine-oriented default; aggregated JSON; lines optional.

#### Load action config file

Prefer CUE (.cue) with schema validation.

#### Locators as file path or URL

#### Map meta records

Transform {locator, meta} -> any.

#### One file per locator

Minimize merge conflicts.

#### Process in parallel

Goroutines + channels; bounded pool; default workers = CPU count.

#### Reduce across meta set

Aggregate stream -> single result.

#### Respect .gitignore by default

Always on; opt-out via --no-gitignore.

#### Run shell using map output

Support bash, sh, zsh early.

#### Script filter/map/reduce

Lua only (v1): small + popular.

#### Update many meta files

#### Validate {locator, meta} schema

Required fields: locator:string, meta:object; error on missing.

#### Validate meta files only

No transforms or shell; emit validation report.

