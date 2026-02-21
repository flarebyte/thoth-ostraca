# Upgrading

## Config version policy

`thoth` currently supports `configVersion: "1"` only.

If your config uses another version, the CLI fails with a short error like:

`unsupported configVersion: "2" (supported: 1)`

## Defaults and user-visible behavior

- Discovery honors `.gitignore` by default.
- Output defaults to machine-oriented compact JSON on stdout.
- Progress output is disabled by default.

## Output modes

- Compact JSON (default): no flag required.
- Pretty JSON: set `output.pretty: true`.
- JSON lines: set `output.lines: true`.
- File output: set `output.out: "path/to/file.json"`.
