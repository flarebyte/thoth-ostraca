# thoth

## What is thoth?
`thoth` is a Go CLI for metadata management in repositories. It discovers files, reads/writes `*.thoth.yaml` metadata, and runs deterministic pipelines for validate/create/update/diff workflows.

It also supports programmable file intelligence with Lua. You can derive expected metadata per locator, transform records, and run controlled shell/Lua stages while keeping machine-friendly JSON outputs stable across runs.

![thoth-ostraca](./thoth-ostraca.png)

## Quickstart
Build locally:

```bash
go build -o .e2e-bin/thoth ./cmd/thoth
```

First run (validate existing meta files):

```bash
cat > config.cue <<'CUE'
{
  configVersion: "1"
  action: "validate"
  discovery: { root: "./repo" }
}
CUE

./.e2e-bin/thoth run --config config.cue
```

## Concepts
- `locator`: canonical path/id for an input file (for example `src/a.txt`).
- `meta file`: YAML sidecar at `<locator>.thoth.yaml` with shape `{ locator, meta }`.
- `actions`:
  - `pipeline`: discover + parse + validate + optional Lua/shell pipeline.
  - `validate`: parse/validate locators only.
  - `create-meta`: create missing meta files.
  - `update-meta`: merge updates into meta files.
  - `diff-meta`: compare existing meta vs expected baseline.
- `stage pipeline`: each action maps to an ordered list of stages.
- `deterministic outputs`: stable sorting and canonical JSON/YAML to keep outputs byte-identical across reruns/workers.

## Common Workflows
### Validate meta files
```cue
// validate.cue
{
  configVersion: "1"
  action: "validate"
  discovery: { root: "./repo" }
}
```

```bash
./.e2e-bin/thoth run --config validate.cue
```

### Create meta files
```cue
// create.cue
{
  configVersion: "1"
  action: "create-meta"
  discovery: { root: "./repo" }
}
```

```bash
./.e2e-bin/thoth run --config create.cue
```

### Update meta files (global patch)
```cue
// update_patch.cue
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

```bash
./.e2e-bin/thoth run --config update_patch.cue
```

### Update meta files (per-locator Lua expected)
```cue
// update_lua.cue
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

```bash
./.e2e-bin/thoth run --config update_lua.cue
```

### Diff meta files (summary / detailed / json-patch)
Summary + drift exit code (`2`) when changed:

```cue
// diff_summary.cue
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

Detailed:

```cue
// diff_detailed.cue
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

JSON Patch (RFC 6902):

```cue
// diff_jsonpatch.cue
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

```bash
./.e2e-bin/thoth run --config diff_summary.cue
./.e2e-bin/thoth run --config diff_detailed.cue
./.e2e-bin/thoth run --config diff_jsonpatch.cue
```

### Diagnose a stage with prepared input
Run the same routed pipeline as `run`, but stop at an intermediate stage:

```bash
./.e2e-bin/thoth diagnose \
  --prepare-pipeline diff-meta \
  --until-stage compute-meta-diff \
  --config diff_summary.cue
```

## Output Modes
Default: compact JSON envelope to `stdout`.

```cue
output: {
  out: "-"      // stdout
  pretty: false  // compact JSON
  lines: false   // envelope mode
}
```

NDJSON record streaming:

```cue
output: {
  out: "-"
  lines: true
}
```

Write to file + pretty JSON:

```cue
output: {
  out: "./out.json"
  pretty: true
}
```

## Safety Notes
- Lua runs inside a sandbox with configurable limits (`meta.luaSandbox`): timeout, instruction limit, memory limit, deterministic random.
- Shell execution is opt-in (`shell.enabled=true`); treat configs as code. Use strict templating and timeouts for safer runs.
- Keep default machine output on `stdout`; diagnostics/progress/summary are emitted to `stderr` only when enabled.

## Repository Layout
- `cmd/thoth/`: CLI entrypoints (`run`, `diagnose`, `version`).
- `internal/stage/`: pipeline stages and stage tests.
- `script/e2e/`: end-to-end tests (TypeScript).
- `testdata/configs/`: config fixtures used in tests.
- `testdata/repos/`: input/meta repositories for scenarios.
- `testdata/run/`: golden JSON outputs.
- `testdata/contracts/`: contract snapshot goldens.
