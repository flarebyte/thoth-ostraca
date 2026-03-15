# Practical Features Gap

## Purpose

This document compares the original intent in
[docs/designs/FLOW_DESIGN.md](/Users/olivier/Documents/github/thoth-ostraca/docs/designs/FLOW_DESIGN.md)
with the current practical behavior of the CLI.

The goal is not to restate every planned feature. The goal is to isolate
the one workflow that should make the CLI useful in practice:

- discover files
- filter them
- run shell analysis per file
- map the results
- optionally reduce the results
- save the results either as:
  - JSON output, or
  - updated `.thoth.yaml` sidecars

JSON, and in some cases YAML, must be treated as first-class data
throughout that workflow.

## Essential Expected Workflow

The design doc describes this as a first-class capability:

- `pipeline`: filter -> map -> shell -> post-map -> reduce -> JSON output
- `create`: discover files -> filter -> map -> post-map -> save meta files
- `update`: discover files -> filter -> map -> post-map with existing ->
  update meta files
- `diff`: discover files -> filter -> map -> post-map with existing ->
  diff existing vs desired

In practical terms, a user should be able to:

1. select files of interest
2. analyse each file with a tool such as `maat-ostraca`
3. treat the tool output as structured JSON
4. transform that structured result into metadata
5. either emit it as output or persist it into `.thoth.yaml`

That is the core utility. Everything else is secondary.

## What The Current CLI Actually Delivers

The shipped implementation supports only part of that design:

- `pipeline` / `nop` works on existing `.thoth.yaml` files, not arbitrary
  input files
- `create-meta` and `update-meta` work on input files, but they do not
  expose `filter`, `map`, `shell`, or `postMap` in the run path
- shell output is captured as strings, not decoded structured JSON
- sidecar writes happen in-place beside source files
- discovery is too broad for practical repo-scale usage without stronger
  defaults and excludes

The result is that the one practically useful workflow is missing:

- file discovery + file filtering + shell analysis + structured mapping +
  persistence

## Main Gap

The design says the CLI is an action-configured data pipeline.

The implementation behaves more like two disconnected products:

- a meta-file pipeline (`pipeline` / `nop`)
- file-sidecar maintenance commands (`create-meta`, `update-meta`,
  `diff-meta`)

The practical bridge between them was never completed.

## Where It Went Off-Rail

### 1. The implementation followed stage names, not user workflows

The repository has many tests around individual stages and deterministic
outputs, but the most important end-to-end workflow was not validated:

- can a user start from source files,
- run shell analysis,
- transform structured output,
- and persist useful metadata?

The implementation became internally testable without becoming externally
useful.

### 2. The shipped behavior drifted from the design docs

`FLOW_DESIGN.md` describes:

- file-oriented `create`
- file-oriented `update`
- filter/map/post-map on files
- save policies and dry-run behavior
- richer shell templating
- JSON-friendly post-shell processing

The shipped CLI does not deliver that set cohesively.

### 3. JSON was not treated as a first-class shell result

This is a critical gap.

In practice, shell analysis tools emit JSON. The CLI should therefore make
JSON shell output easy to consume.

Today, shell output is exposed as raw strings only. That forces awkward
workarounds:

- extra helper commands such as `jq`
- fragile string parsing in Lua
- complicated `postMap` code

That is the opposite of the intended data-pipeline experience.

### 4. File persistence was treated separately from analysis

`create-meta` and `update-meta` persist sidecars, but they do not expose
the programmable analysis path that users need.

`pipeline` exposes the programmable path, but it does not persist sidecars.

This split makes the CLI hard to use for real metadata enrichment tasks.

### 5. Safety and practicality defaults were not validated against real repos

Observed practical failures:

- creating sidecars under `.git`
- colliding with fixture directories under `internal`
- failing on existing sidecars without a practical ignore/update mode
- no dedicated output directory for generated sidecars
- no user-visible progress during long-running actions

These are not minor polish issues. They block normal usage.

## Practical Must-Haves To Make The CLI Useful

The CLI needs one first-class workflow to be spec'd and protected:

- discover input files
- filter input files
- run shell analysis per file
- decode shell JSON automatically
- map decoded results into metadata
- optionally reduce results
- either:
  - write JSON output, or
  - write/update `.thoth.yaml` sidecars

Supporting requirements:

- discovery must exclude `.git` and similar internal directories by default
- generated sidecars must be able to go into a dedicated output directory
- save mode must support dry-run
- shell templating must support direct field interpolation, not only a
  single `{json}` placeholder
- long-running actions must expose progress
- Lua/scripting limits must not reject ordinary transformation logic by
  default

## Practical Guardrails For AI-Built Projects

### 1. Define one critical user journey before implementation

Before writing code, pick one concrete scenario and keep it executable from
day one.

For this project, the scenario should have been:

- choose Go files
- run shell analysis per file
- transform output into metadata
- save results

If that journey is not working, the project is not done.

### 2. Require one acceptance test per top-level promise

Do not stop at unit tests and stage-level golden files.

Every user-facing promise in the design must have at least one real
acceptance test:

- file discovery to output
- file discovery to sidecar save
- shell JSON to mapped metadata
- update existing sidecars from computed results

### 3. Track design-to-implementation drift explicitly

The moment the implementation renames or narrows a feature, record it in a
living gap document.

Do not let generated docs continue describing behavior that the shipped CLI
does not actually support.

### 4. Prefer capability tests over structural tests

Good:

- "can analyse 100 files and write metadata"

Weak:

- "stage X emits shape Y"

Structural tests are useful, but they cannot replace capability tests.

### 5. Treat data interchange formats as first-class

If the product expects shell tools and AI/CLI automation:

- JSON decoding must be native
- YAML writing must be deterministic and configurable
- text parsing should be a fallback, not the primary integration method

### 6. Fail early on unsafe defaults

If a command would write into:

- `.git`
- fixture directories
- broad roots without excludes

the CLI should warn or refuse by default.

## Short Conclusion

The design aimed at a practical metadata-processing pipeline.

The shipped CLI mostly delivered stage machinery and deterministic test
coverage, but it missed the single workflow that would make the tool useful
in practice:

- file -> filter -> shell -> structured map -> save

The immediate spec priority is therefore not more add-on features.
It is to restore that core workflow as a protected, acceptance-tested,
user-visible capability.
