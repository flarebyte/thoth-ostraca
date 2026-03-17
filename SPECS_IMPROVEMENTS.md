# Specs Improvements

## Must-Have

- `create-meta` must support writing generated `.thoth.yaml` files to a dedicated output directory instead of only alongside source files.
- The output directory must preserve source-relative paths so generated metadata stays deterministic and collision-free.
- The feature must support a dry-run mode that reports planned writes without modifying the repository.
- Input-file discovery must exclude VCS and tool-internal directories such as `.git` by default.
- `thoth run` must support a first-class input-file pipeline that can discover files, filter them, map them, run shell analysis per file, and persist the computed metadata back into `.thoth.yaml` sidecars.
- Long-running actions such as `pipeline`, `create-meta`, and `update-meta` must expose an explicit progress flag so users can see stage-level and record-level progress while work is running.
- Lua sandbox instruction-limit handling must allow ordinary post-processing loops used to transform shell output, or provide a clear configurable way to raise that limit safely.
- Shell stage output must support automatic JSON decoding so downstream steps can consume structured results without extra helper commands such as `jq`.
- Shell templating must support direct field interpolation for common values such as `locator` and mapped fields, not only a single `{json}` placeholder.
