# thoth Cheatsheet

Recipes-first quick reference. See `README.md` for full concepts.

## Action Matrix

| Action             | Discovers                         | Filter | Map | Shell | PostMap | Reduce | Persist sidecars | JSON output |
| ------------------ | --------------------------------- | -----: | --: | ----: | ------: | -----: | ---------------: | ----------: |
| `pipeline` / `nop` | existing `.thoth.yaml` meta files |    yes | yes |   yes |     yes |    yes |               no |         yes |
| `input-pipeline`   | arbitrary input files             |    yes | yes |   yes |     yes |    yes |              yes |         yes |
| `validate`         | existing `.thoth.yaml` meta files |     no |  no |    no |      no |     no |               no |         yes |
| `create-meta`      | arbitrary input files             |    yes |  no |    no |      no |     no |              yes |         yes |
| `update-meta`      | arbitrary input files             |    yes |  no |    no |      no |     no |              yes |         yes |
| `diff-meta`        | input files + existing meta files |    yes |  no |    no |      no |     no |               no |         yes |

Notes:

- `pipeline` / `nop` is the programmable meta-file workflow. It works on
  existing sidecars, not arbitrary source files.
- `input-pipeline` is the practical file workflow. It supports discovery,
  filter, map, shell, postMap, reduce, JSON output, progress, dry-run
  persistence, and dedicated sidecar output directories.
- `create-meta` and `update-meta` are narrower metadata maintenance actions.
  `create-meta` now supports `lua-filter`, but not `map`, `shell`, or `reduce`.
  `update-meta` now supports `lua-filter`, but not `map`, `shell`, or `reduce`.
- `diff-meta` now supports `lua-filter` on input files. Orphan meta reporting
  still reflects the full discovered meta-file set under the chosen root, so
  filtered-out paired files can become reported as orphans.

## Top 10 Commands / Workflows

```bash
# 1) Build local binary
go build -o .e2e-bin/thoth ./cmd/thoth

# 2) Run with config
./.e2e-bin/thoth run --config ./config.cue

# 3) Validate existing meta files
./.e2e-bin/thoth run --config ./validate.cue

# 4) Create missing .thoth.yaml files
./.e2e-bin/thoth run --config ./create_meta.cue

# 5) Update meta files (global patch)
./.e2e-bin/thoth run --config ./update_patch.cue

# 6) Update meta files (per-locator Lua)
./.e2e-bin/thoth run --config ./update_lua.cue

# 7) Diff meta (summary)
./.e2e-bin/thoth run --config ./diff_summary.cue

# 8) Diff meta (detailed / json-patch)
./.e2e-bin/thoth run --config ./diff_detailed.cue
./.e2e-bin/thoth run --config ./diff_jsonpatch.cue

# 9) Diagnose a stage directly
./.e2e-bin/thoth diagnose --stage validate-config --config ./config.cue

# 10) Run tests quickly
go test ./...
```

## Config Snippets (CUE)

### validate

```cue
{
  configVersion: "1"
  action: "validate"
  discovery: { root: "./repo" }
}
```

### create-meta

```cue
{
  configVersion: "1"
  action: "create-meta"
  discovery: { root: "./repo" }
}
```

### update-meta with patch

```cue
{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "./repo" }
  updateMeta: {
    patch: {
      owner: "team-a"
      tags: ["managed"]
      nested: { reviewed: true }
    }
  }
}
```

### update-meta with expectedLua

```cue
{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "./repo" }
  updateMeta: {
    expectedLua: {
      inline: '''
return function(locator, existingMeta)
  if locator == "a.txt" then
    return { priority = "high" }
  end
  return { priority = "normal" }
end
'''
    }
  }
}
```

### diff-meta summary

```cue
{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "./repo" }
  diffMeta: {
    format: "summary"
    expectedPatch: { owner: "team-a" }
  }
}
```

### diff-meta detailed

```cue
{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "./repo" }
  diffMeta: {
    format: "detailed"
    expectedPatch: { owner: "team-a" }
  }
}
```

### diff-meta json-patch

```cue
{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "./repo" }
  diffMeta: {
    format: "json-patch"
    expectedPatch: { owner: "team-a" }
  }
}
```

### diff-meta only=changed

```cue
{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "./repo" }
  diffMeta: {
    only: "changed"
    expectedPatch: { owner: "team-a" }
  }
}
```

### diff-meta failOnChange

```cue
{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "./repo" }
  diffMeta: {
    format: "summary"
    failOnChange: true
    expectedPatch: { owner: "team-a" }
  }
}
```

### output.lines=true (streaming NDJSON)

```cue
{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "./repo" }
  output: {
    out: "-"
    lines: true
  }
}
```

## Diagnose Recipes

### Prepare input-files/meta-files

```bash
# Prepare input-files then run a stage
./.e2e-bin/thoth diagnose \
  --prepare input-files \
  --root ./repo \
  --stage discover-input-files

# Prepare meta-files then run a stage
./.e2e-bin/thoth diagnose \
  --prepare meta-files \
  --root ./repo \
  --stage discover-meta-files
```

### Prepare full action pipeline and run until stage

```bash
./.e2e-bin/thoth diagnose \
  --prepare-pipeline diff-meta \
  --until-stage compute-meta-diff \
  --config ./diff_summary.cue

# Equivalent by index (0-based)
./.e2e-bin/thoth diagnose \
  --prepare-pipeline diff-meta \
  --until-index 4 \
  --config ./diff_summary.cue
```

### Dump fixtures to a directory

```bash
./.e2e-bin/thoth diagnose \
  --prepare-pipeline pipeline \
  --until-stage parse-validate-yaml \
  --dump-dir ./tmp/diag \
  --config ./config.cue
```

## Troubleshooting Quick Hits

```text
invalid locator
- Check locator policy (absolute paths, parent refs, backslashes, URLs).

missing required field: meta / YAML schema errors
- Each .thoth.yaml must contain:
  locator: <string>
  meta: <object>

lua-*: sandbox timeout / instruction / memory
- Reduce Lua complexity or increase luaSandbox limits in config.

shell-exec: strict templating: invalid placeholder
- With strict templating, only supported placeholders are allowed (e.g. {json}).
```

## Useful Test Commands

```bash
go test ./...
make test-race
make bench
```
