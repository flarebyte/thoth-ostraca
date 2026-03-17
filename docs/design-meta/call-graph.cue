package flyb

#callGraphNotes: [
  {
    name:   "call.thoth.cli.root"
    title:  "thoth CLI root command"
    labels: ["call", "design", "flow"]
    markdown: """
- note: cobra-based command tree
- pkg: cmd/thoth
- func: CliRoot
- file: cmd/thoth/cli_root.go
"""
  },
  {
    name:   "call.run.parse-args"
    title:  "Parse args for run"
    labels: ["call", "design", "flow"]
    markdown: """
- note: flags: --config (CUE .cue file). All other options belong in the
  action config.
- pkg: cmd/thoth
- func: CliRun
- file: cmd/thoth/cli_run.go
"""
  },
  {
    name:   "call.run.load-action-config-file"
    title:  "Load action config file"
    labels: ["call", "design", "flow"]
    markdown: """
- note: --config path; CUE schema-validated .cue; drives entire pipeline
- pkg: internal/config
- func: ActionConfigLoad
- file: internal/config/config_load.go
"""
  },
  {
    name:   "call.run.route-by-action-type"
    title:  "Route by action type"
    labels: ["call", "design", "flow"]
    markdown: """
- note: action: pipeline | create | update | diff
- pkg: internal/config
- func: ActionRoute
- file: internal/config/action_route.go
"""
  },
  {
    name:   "call.pipeline.meta-flow"
    title:  "Meta pipeline flow"
    labels: ["call", "design", "flow"]
    markdown: """
- pkg: internal/pipeline
- func: FlowPipeline
- file: internal/pipeline/flow_pipeline.go
"""
  },
  {
    name:   "call.pipeline.find-meta-files"
    title:  "Find *.thoth.yaml files"
    labels: ["call", "design", "flow"]
    markdown: """
- note: walk root; .gitignore ON by default even outside git repos;
  --no-gitignore to disable; do not follow symlinks by default
- pkg: internal/fs
- func: FsDiscovery
- file: internal/fs/fs_discovery.go
"""
  },
  {
    name:   "call.pipeline.parse-validate-yaml"
    title:  "Parse and validate YAML records"
    labels: ["call", "design", "flow"]
    markdown: """
- note: yaml.v3; strict fields; types; locator canonicalization; top-level
  unknown = error (unless validation.allowUnknownTopLevel); inside meta:
  unknown allowed
- pkg: internal/meta
- func: MetaParse
- file: internal/meta/meta_parse.go
"""
  },
  {
    name:   "call.pipeline.apply-filter"
    title:  "Apply filter predicate"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only predicate (v1)
- pkg: internal/pipeline
- func: MetaFilterStep
- file: internal/pipeline/meta_filter_step.go
"""
  },
  {
    name:   "call.pipeline.apply-map"
    title:  "Apply map transform"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only mapping (v1); parallel by default
- pkg: internal/pipeline
- func: MetaMapStep
- file: internal/pipeline/meta_map_step.go
"""
  },
  {
    name:   "call.pipeline.execute-shell"
    title:  "Execute shell per mapped item"
    labels: ["call", "design", "flow"]
    markdown: """
- note: conditional shell execution; argv templates preferred; timeout kills
  process group
- pkg: internal/shell
- func: ShellExec
- file: internal/shell/shell_exec.go
"""
  },
  {
    name:   "call.pipeline.post-map-shell-results"
    title:  "Post-map shell results"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua transforms {locator,input,shell:{cmd,exitCode,stdout,stderr,
  durationMs}}
- pkg: internal/pipeline
- func: MetaMapPostShell
- file: internal/pipeline/meta_post_shell.go
"""
  },
  {
    name:   "call.pipeline.apply-reduce"
    title:  "Apply reduce aggregate"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only reduce (v1); parallel feed; single JSON value
- pkg: internal/pipeline
- func: MetaReduceStep
- file: internal/pipeline/meta_reduce_step.go
"""
  },
  {
    name:   "call.pipeline.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: aggregated JSON array by default; --lines streams nondeterministically;
  reduce -> single value
- pkg: internal/output
- func: OutputJsonResult
- file: internal/output/json_result.go
"""
  },
  {
    name:   "call.create.flow"
    title:  "Create meta files flow"
    labels: ["call", "design", "flow"]
    markdown: """
- pkg: internal/pipeline
- func: FlowCreate
- file: internal/pipeline/flow_create.go
"""
  },
  {
    name:   "call.create.find-files"
    title:  "Find files recursively (gitignore)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: walk root; .gitignore ON by default; no patterns; do not follow
  symlinks by default; filenames as inputs
- pkg: internal/fs
- func: FsDiscoveryFiles
- file: internal/fs/discovery_files.go
"""
  },
  {
    name:   "call.create.enrich-files"
    title:  "Enrich files with OS/Git info"
    labels: ["call", "design", "flow"]
    markdown: """
- note: conditional files.info and/or files.git enrichment
- pkg: internal/pipeline
- func: FilesEnrich
- file: internal/pipeline/files_enrich.go
"""
  },
  {
    name:   "call.create.filter-filenames"
    title:  "Filter filenames"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only predicate (v1) over {file}
- pkg: internal/pipeline
- func: FilesFilterStep
- file: internal/pipeline/files_filter_step.go
"""
  },
  {
    name:   "call.create.map-filenames"
    title:  "Map filenames"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only map (v1) over {file}
- pkg: internal/pipeline
- func: FilesMapStep
- file: internal/pipeline/files_map_step.go
"""
  },
  {
    name:   "call.create.post-map-from-files"
    title:  "Post-map from files"
    labels: ["call", "design", "flow"]
    markdown: """
- note: conditional inline Lua transforms {file,input} -> any
- pkg: internal/pipeline
- func: FilesMapPost
- file: internal/pipeline/files_map_post.go
"""
  },
  {
    name:   "call.create.save-meta-files"
    title:  "Save meta files (*.thoth.yaml)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: conditional save; naming convention based on sha256 root+relPath hash
- pkg: internal/save
- func: MetaSave
- file: internal/save/meta_save.go
"""
  },
  {
    name:   "call.create.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: aggregated JSON array by default; --lines streams nondeterministically;
  reduce -> single value
- pkg: internal/output
- func: OutputJsonResult
- file: internal/output/json_result.go
"""
  },
  {
    name:   "call.update.flow"
    title:  "Update meta files flow"
    labels: ["call", "design", "flow"]
    markdown: """
- pkg: internal/pipeline
- func: FlowUpdate
- file: internal/pipeline/flow_update.go
"""
  },
  {
    name:   "call.update.find-files"
    title:  "Find files recursively (update)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: walk root; .gitignore ON by default; do not follow symlinks by default
- pkg: internal/fs
- func: FsDiscoveryFilesUpdate
- file: internal/fs/files_update.go
"""
  },
  {
    name:   "call.update.enrich-files"
    title:  "Enrich files with OS/Git info"
    labels: ["call", "design", "flow"]
    markdown: """
- note: conditional files.info and/or files.git enrichment
- pkg: internal/pipeline
- func: FilesEnrich
- file: internal/pipeline/files_enrich.go
"""
  },
  {
    name:   "call.update.filter-filenames"
    title:  "Filter filenames"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only predicate (v1) over {file}
- pkg: internal/pipeline
- func: FilesFilterStep
- file: internal/pipeline/files_filter_step.go
"""
  },
  {
    name:   "call.update.map-filenames"
    title:  "Map filenames"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only map (v1) over {file}
- pkg: internal/pipeline
- func: FilesMapStep
- file: internal/pipeline/files_map_step.go
"""
  },
  {
    name:   "call.update.load-existing-meta"
    title:  "Load existing meta (if any)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: compute expected path by naming convention; read YAML if exists
- pkg: internal/meta
- func: MetaLoadExisting
- file: internal/meta/load_existing.go
"""
  },
  {
    name:   "call.update.post-map-for-update"
    title:  "Post-map for update (with existing)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua receives {file,input,existing?}; returns either { meta } or
  { patch } (RFC6902)
- pkg: internal/pipeline
- func: FilesMapPostUpdate
- file: internal/pipeline/files_post_update.go
"""
  },
  {
    name:   "call.update.update-meta-files"
    title:  "Update meta files (merge/create)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: shallow|deep|jsonpatch merge strategy; missing entries can create new
  sidecars
- pkg: internal/save
- func: MetaUpdate
- file: internal/save/meta_update.go
"""
  },
  {
    name:   "call.update.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: aggregated JSON array by default; --lines streams nondeterministically;
  reduce -> single value
- pkg: internal/output
- func: OutputJsonResult
- file: internal/output/json_result.go
"""
  },
  {
    name:   "call.diff.flow"
    title:  "Diff meta files flow"
    labels: ["call", "design", "flow"]
    markdown: """
- pkg: internal/pipeline
- func: FlowDiff
- file: internal/pipeline/flow_diff.go
"""
  },
  {
    name:   "call.diff.find-files"
    title:  "Find files recursively (update)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: walk root; .gitignore ON by default; do not follow symlinks by default
- pkg: internal/fs
- func: FsDiscoveryFilesUpdate
- file: internal/fs/files_update.go
"""
  },
  {
    name:   "call.diff.enrich-files"
    title:  "Enrich files with OS/Git info"
    labels: ["call", "design", "flow"]
    markdown: """
- note: conditional files.info and/or files.git enrichment
- pkg: internal/pipeline
- func: FilesEnrich
- file: internal/pipeline/files_enrich.go
"""
  },
  {
    name:   "call.diff.filter-filenames"
    title:  "Filter filenames"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only predicate (v1) over {file}
- pkg: internal/pipeline
- func: FilesFilterStep
- file: internal/pipeline/files_filter_step.go
"""
  },
  {
    name:   "call.diff.map-filenames"
    title:  "Map filenames"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua-only map (v1) over {file}
- pkg: internal/pipeline
- func: FilesMapStep
- file: internal/pipeline/files_map_step.go
"""
  },
  {
    name:   "call.diff.load-existing-meta"
    title:  "Load existing meta (if any)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: compute expected path by naming convention; read YAML if exists
- pkg: internal/meta
- func: MetaLoadExisting
- file: internal/meta/load_existing.go
"""
  },
  {
    name:   "call.diff.post-map-for-update"
    title:  "Post-map for update (with existing)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: Lua receives {file,input,existing?}; returns either { meta } or
  { patch } (RFC6902)
- pkg: internal/pipeline
- func: FilesMapPostUpdate
- file: internal/pipeline/files_post_update.go
"""
  },
  {
    name:   "call.diff.compute-meta-diffs"
    title:  "Compute meta diffs"
    labels: ["call", "design", "flow"]
    markdown: """
- note: deep diff existing vs patch-applied result; output RFC6902 JSON Patch
  + summary
- pkg: internal/diff
- func: MetaDiffCompute
- file: internal/diff/diff_compute.go
"""
  },
  {
    name:   "call.diff.detect-orphans"
    title:  "Detect orphan meta files"
    labels: ["call", "design", "flow"]
    markdown: """
- note: iterate *.thoth.yaml; if locator is file path and does not exist, flag
- pkg: internal/diff
- func: MetaDiffOrphans
- file: internal/diff/diff_orphans.go
"""
  },
  {
    name:   "call.diff.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: aggregated JSON array by default; --lines streams nondeterministically;
  reduce -> single value
- pkg: internal/output
- func: OutputJsonResult
- file: internal/output/json_result.go
"""
  },
  {
    name:   "call.validate.flow"
    title:  "Validate meta files only"
    labels: ["call", "design", "flow"]
    markdown: """
- pkg: internal/pipeline
- func: FlowValidate
- file: internal/pipeline/flow_validate.go
"""
  },
  {
    name:   "call.validate.find-meta-files"
    title:  "Find *.thoth.yaml files"
    labels: ["call", "design", "flow"]
    markdown: """
- note: walk root; .gitignore ON by default even outside git repos;
  --no-gitignore to disable; do not follow symlinks by default
- pkg: internal/fs
- func: FsDiscovery
- file: internal/fs/fs_discovery.go
"""
  },
  {
    name:   "call.validate.parse-validate-yaml"
    title:  "Parse and validate YAML records"
    labels: ["call", "design", "flow"]
    markdown: """
- note: yaml.v3; strict fields; types; locator canonicalization; top-level
  unknown = error (unless validation.allowUnknownTopLevel); inside meta:
  unknown allowed
- pkg: internal/meta
- func: MetaParse
- file: internal/meta/meta_parse.go
"""
  },
  {
    name:   "call.validate.collect-validation-results"
    title:  "Collect validation results"
    labels: ["call", "design", "flow"]
    markdown: """
- note: schema + locator checks only; no filter/map/reduce/shell
- pkg: internal/pipeline
- func: MetaValidateOnly
- file: internal/pipeline/validate_only.go
"""
  },
  {
    name:   "call.validate.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: aggregated JSON array by default; --lines streams nondeterministically;
  reduce -> single value
- pkg: internal/output
- func: OutputJsonResult
- file: internal/output/json_result.go
"""
  },
  {
    name:   "call.diagnose.parse-args"
    title:  "Parse args for diagnose"
    labels: ["call", "design", "flow"]
    markdown: """
- note: diagnose subcommand: --config, --step, input selection flags, dump
  flags, debug flags
- pkg: cmd/thoth
- func: CliDiagnose
- file: cmd/thoth/cli_diagnose.go
"""
  },
  {
    name:   "call.diagnose.single-stage"
    title:  "Diagnose single stage"
    labels: ["call", "design", "flow"]
    markdown: """
- pkg: internal/pipeline
- func: FlowDiagnose
- file: internal/pipeline/flow_diagnose.go
"""
  },
  {
    name:   "call.diagnose.parse-subcommand-args"
    title:  "Parse args for diagnose"
    labels: ["call", "design", "flow"]
    markdown: """
- note: --config, --step, --input-file|--input-inline|--input-stdin,
  --dump-in, --dump-out, --limit, --seed, --dry-shell
- pkg: internal
- func: DiagnoseParseArgs
- file: internal/parse_args.go
"""
  },
  {
    name:   "call.diagnose.load-action-config"
    title:  "Load action config (CUE)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: use existing action config; validate with CUE schema
- pkg: internal/config
- func: DiagnoseConfigLoad
- file: internal/config/config_load.go
"""
  },
  {
    name:   "call.diagnose.resolve-target-step"
    title:  "Resolve target step"
    labels: ["call", "design", "flow"]
    markdown: """
- note: map stable step name to internal implementation based on action
- pkg: internal
- func: DiagnoseStepResolve
- file: internal/step_resolve.go
"""
  },
  {
    name:   "call.diagnose.resolve-input-mode"
    title:  "Resolve input mode"
    labels: ["call", "design", "flow"]
    markdown: """
- note: use explicit JSON or prepare upstream to boundary; apply --limit/--seed
- pkg: internal
- func: DiagnoseInputResolve
- file: internal/input_resolve.go
"""
  },
  {
    name:   "call.diagnose.dump-stage-input"
    title:  "Dump stage input (optional)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: --dump-in [path|-]; emit boundary input as JSON/NDJSON
- pkg: internal
- func: DiagnoseDumpIn
- file: internal/dump_in.go
"""
  },
  {
    name:   "call.diagnose.emit-run-header"
    title:  "Emit run header"
    labels: ["call", "design", "flow"]
    markdown: """
- note: structured log: { action, executedStep, preparedStages, inputMode,
  limits }
- pkg: internal
- func: DiagnoseHeaderEmit
- file: internal/header_emit.go
"""
  },
  {
    name:   "call.diagnose.execute-target-stage"
    title:  "Execute target stage"
    labels: ["call", "design", "flow"]
    markdown: """
- note: run only the selected step; --dry-shell renders command/env without
  exec for shell stage
- pkg: internal
- func: DiagnoseStageExec
- file: internal/stage_exec.go
"""
  },
  {
    name:   "call.diagnose.dump-stage-output"
    title:  "Dump stage output (optional)"
    labels: ["call", "design", "flow"]
    markdown: """
- note: --dump-out [path|-]; emit stage output boundary for reproducible
  debugging
- pkg: internal
- func: DiagnoseDumpOut
- file: internal/dump_out.go
"""
  },
]

#callGraphRelationships: [
  { from: "call.thoth.cli.root", to: "call.run.parse-args", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.parse-args", to: "call.run.load-action-config-file", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.parse-args", to: "call.run.route-by-action-type", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.route-by-action-type", to: "call.pipeline.meta-flow", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.find-meta-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.parse-validate-yaml", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.apply-filter", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.apply-map", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.execute-shell", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.post-map-shell-results", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.apply-reduce", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.pipeline.meta-flow", to: "call.pipeline.write-json-result", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.route-by-action-type", to: "call.create.flow", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.find-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.enrich-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.filter-filenames", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.map-filenames", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.post-map-from-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.save-meta-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.create.flow", to: "call.create.write-json-result", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.route-by-action-type", to: "call.update.flow", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.find-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.enrich-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.filter-filenames", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.map-filenames", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.load-existing-meta", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.post-map-for-update", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.update-meta-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.update.flow", to: "call.update.write-json-result", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.route-by-action-type", to: "call.diff.flow", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.find-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.enrich-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.filter-filenames", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.map-filenames", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.load-existing-meta", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.post-map-for-update", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.compute-meta-diffs", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.detect-orphans", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diff.flow", to: "call.diff.write-json-result", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.run.route-by-action-type", to: "call.validate.flow", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.validate.flow", to: "call.validate.find-meta-files", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.validate.flow", to: "call.validate.parse-validate-yaml", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.validate.flow", to: "call.validate.collect-validation-results", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.validate.flow", to: "call.validate.write-json-result", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.thoth.cli.root", to: "call.diagnose.parse-args", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.parse-args", to: "call.diagnose.single-stage", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.parse-subcommand-args", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.load-action-config", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.resolve-target-step", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.resolve-input-mode", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.dump-stage-input", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.emit-run-header", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.execute-target-stage", label: "delegate_to", labels: ["delegate_to"] },
  { from: "call.diagnose.single-stage", to: "call.diagnose.dump-stage-output", label: "delegate_to", labels: ["delegate_to"] },
]

#callDetailRefs: [
  "call.thoth.cli.root",
  "call.run.parse-args",
  "call.run.load-action-config-file",
  "call.run.route-by-action-type",
  "call.pipeline.meta-flow",
  "call.pipeline.find-meta-files",
  "call.pipeline.parse-validate-yaml",
  "call.pipeline.apply-filter",
  "call.pipeline.apply-map",
  "call.pipeline.execute-shell",
  "call.pipeline.post-map-shell-results",
  "call.pipeline.apply-reduce",
  "call.pipeline.write-json-result",
  "call.create.flow",
  "call.create.find-files",
  "call.create.enrich-files",
  "call.create.filter-filenames",
  "call.create.map-filenames",
  "call.create.post-map-from-files",
  "call.create.save-meta-files",
  "call.create.write-json-result",
  "call.update.flow",
  "call.update.find-files",
  "call.update.enrich-files",
  "call.update.filter-filenames",
  "call.update.map-filenames",
  "call.update.load-existing-meta",
  "call.update.post-map-for-update",
  "call.update.update-meta-files",
  "call.update.write-json-result",
  "call.diff.flow",
  "call.diff.find-files",
  "call.diff.enrich-files",
  "call.diff.filter-filenames",
  "call.diff.map-filenames",
  "call.diff.load-existing-meta",
  "call.diff.post-map-for-update",
  "call.diff.compute-meta-diffs",
  "call.diff.detect-orphans",
  "call.diff.write-json-result",
  "call.validate.flow",
  "call.validate.find-meta-files",
  "call.validate.parse-validate-yaml",
  "call.validate.collect-validation-results",
  "call.validate.write-json-result",
  "call.diagnose.parse-args",
  "call.diagnose.single-stage",
  "call.diagnose.parse-subcommand-args",
  "call.diagnose.load-action-config",
  "call.diagnose.resolve-target-step",
  "call.diagnose.resolve-input-mode",
  "call.diagnose.dump-stage-input",
  "call.diagnose.emit-run-header",
  "call.diagnose.execute-target-stage",
  "call.diagnose.dump-stage-output",
]
