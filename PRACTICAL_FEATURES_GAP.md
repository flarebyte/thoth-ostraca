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

## What The Current CLI Now Delivers

The shipped implementation now delivers the practical core workflow:

- `input-pipeline` works on arbitrary input files, not only existing
  `.thoth.yaml` sidecars
- the run path supports `discover -> filter -> map -> shell -> postMap -> reduce`
- shell stdout can be decoded as structured JSON
- mapped metadata can be written back to `.thoth.yaml` sidecars
- sidecars can go to a dedicated output directory
- persistence supports dry-run
- discovery excludes unsafe/internal paths by default
- long-running practical workflows can emit progress on stderr
- ordinary Lua transform loops are allowed by default while explicit limits
  are still enforced

The result is that the one practically useful workflow is now restored:

- file discovery + file filtering + shell analysis + structured mapping +
  persistence

## Remaining Gaps

The remaining gaps are narrower than before:

- `create-meta` and `update-meta` are still separate maintenance actions
  instead of thin wrappers over the restored programmable file pipeline
- `diff-meta` still has narrower scoped semantics than the practical
  programmable pipeline, especially around orphan reporting after input
  filtering
- some original design ambitions remain broader than the shipped CLI, but
  the practical critical workflow is no longer missing

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

This was a critical gap and is now fixed.

In practice, shell analysis tools emit JSON. The CLI should therefore make
JSON shell output easy to consume.

Earlier, shell output was exposed as raw strings only. That forced awkward
workarounds:

- extra helper commands such as `jq`
- fragile string parsing in Lua
- complicated `postMap` code

The current CLI now supports opt-in JSON decoding, which is much closer to
the intended data-pipeline experience.

### 4. File persistence was treated separately from analysis

This was another critical split and is now largely fixed for the practical
workflow.

`input-pipeline` now exposes the programmable path and can also persist
sidecars.

`create-meta` and `update-meta` still exist as narrower maintenance actions,
which is now more of an action-surface simplification issue than a core
capability gap.

### 5. Safety and practicality defaults were not validated against real repos

Observed practical failures that drove the gap analysis were:

- creating sidecars under `.git`
- colliding with fixture directories under `internal`
- failing on existing sidecars without a practical ignore/update mode
- no dedicated output directory for generated sidecars
- no user-visible progress during long-running actions

These were not minor polish issues. They blocked normal usage until the
practical file pipeline was restored.

## Practical Must-Haves That Made The CLI Useful

The CLI needs one first-class workflow to be spec'd and protected:

- discover input files
- filter input files
- run shell analysis per file
- decode shell JSON automatically
- map decoded results into metadata
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

These capabilities are now shipped and covered by acceptance tests.

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

The project originally delivered stage machinery and deterministic test
coverage before it delivered the single workflow that would make the tool
useful in practice:

- file -> filter -> shell -> structured map -> save

That core workflow is now restored and acceptance-tested.
The remaining design work is to simplify and unify the surrounding action
surface rather than to recover basic usefulness.
