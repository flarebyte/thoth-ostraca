# Usage

## `create-meta`

Use `create-meta` to bootstrap `.thoth.yaml` sidecar files for discovered input files.

Run it with:

```bash
thoth run --config ./create-meta-project.thoth.cue
```

What it does:

- discovers input files under `discovery.root`
- skips files already ending in `.thoth.yaml`
- creates `<locator>.thoth.yaml` alongside each discovered file
- writes a JSON envelope to `output.out`

What it cannot do:

- it cannot filter discovered input files with `lua-filter`
- it cannot map records with `lua-map`
- it cannot run `shell-exec`
- it cannot write sidecars to a separate output folder
- it cannot be rerun safely on the same files; it fails when a sidecar already exists

Practical consequence:

- if `discovery.root` is too broad, it will try to create sidecars across the whole tree, including fixture directories and other unwanted paths

## `update-meta`

Use `update-meta` to modify existing `.thoth.yaml` sidecar files.

Run it with:

```bash
thoth run --config ./update-meta-project.thoth.cue
```

What it does:

- discovers input files under `discovery.root`
- loads existing sidecars when present
- merges configured metadata changes
- writes updated `.thoth.yaml` files back alongside the source files
- writes a JSON envelope to `output.out`

What it can modify:

- the `meta` content inside the sidecar
- via `updateMeta.patch`
- via `updateMeta.expectedLua.inline`

What it cannot do:

- it cannot filter discovered input files with `lua-filter`
- it cannot map records with `lua-map`
- it cannot run `shell-exec`
- it cannot write sidecars to a separate output folder

## Current Limitation

The shipped CLI does not currently provide a `thoth run` action for:

- discovering arbitrary input files
- filtering them
- mapping them
- optionally running shell commands
- saving the final results somewhere

That is the main gap between the current metadata actions and the broader file-processing workflow.
