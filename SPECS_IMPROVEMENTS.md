# Specs Improvements

## Completed Must-Haves

- `thoth run` now supports a first-class `input-pipeline` action that can discover files, filter them, map them, run shell analysis per file, and persist computed metadata into `.thoth.yaml` sidecars.
- Shell stage output supports opt-in JSON decoding via `shell.decodeJsonStdout`, so downstream steps can consume structured results without helper commands such as `jq`.
- Shell templating supports direct field interpolation for practical placeholders such as `locator`, `file.*`, and `mapped.*` while preserving `{json}`.
- Generated sidecars can be written to a dedicated output directory with deterministic source-relative paths.
- The persistence path supports `persistMeta.dryRun` to preview intended writes without modifying the repository.
- Input-file discovery excludes `.git` and other internal/tooling directories by default, with explicit include/exclude controls.
- Long-running practical workflows can emit explicit progress on stderr via `ui.progress`.
- Lua sandbox instruction-limit handling now allows ordinary transformation loops by default and still enforces explicit configured limits at runtime.

## Remaining Improvement Areas

- `create-meta` and `update-meta` remain narrower metadata maintenance actions rather than aliases over the practical programmable pipeline.
- `diff-meta` still has narrower semantics than the full programmable pipeline; it supports `lua-filter` on input files, but orphan meta reporting still reflects the full discovered meta-file set under the chosen root.
