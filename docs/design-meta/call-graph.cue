package flyb

#callGraphNotes: [
  {
    name:   "call.thoth.cli.root"
    title:  "thoth CLI root command"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.run.parse-args"
    title:  "Parse args for run"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.run.load-action-config-file"
    title:  "Load action config file"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.run.route-by-action-type"
    title:  "Route by action type"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.meta-flow"
    title:  "Meta pipeline flow"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.find-meta-files"
    title:  "Find *.thoth.yaml files"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.parse-validate-yaml"
    title:  "Parse and validate YAML records"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.apply-filter"
    title:  "Apply filter predicate"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.apply-map"
    title:  "Apply map transform"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.execute-shell"
    title:  "Execute shell per mapped item"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.post-map-shell-results"
    title:  "Post-map shell results"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.apply-reduce"
    title:  "Apply reduce aggregate"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.pipeline.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.flow"
    title:  "Create meta files flow"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.find-files"
    title:  "Find files recursively (gitignore)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.enrich-files"
    title:  "Enrich files with OS/Git info"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.filter-filenames"
    title:  "Filter filenames"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.map-filenames"
    title:  "Map filenames"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.post-map-from-files"
    title:  "Post-map from files"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.save-meta-files"
    title:  "Save meta files (*.thoth.yaml)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.create.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.flow"
    title:  "Update meta files flow"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.find-files"
    title:  "Find files recursively (update)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.enrich-files"
    title:  "Enrich files with OS/Git info"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.filter-filenames"
    title:  "Filter filenames"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.map-filenames"
    title:  "Map filenames"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.load-existing-meta"
    title:  "Load existing meta (if any)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.post-map-for-update"
    title:  "Post-map for update (with existing)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.update-meta-files"
    title:  "Update meta files (merge/create)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.update.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.flow"
    title:  "Diff meta files flow"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.find-files"
    title:  "Find files recursively (update)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.enrich-files"
    title:  "Enrich files with OS/Git info"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.filter-filenames"
    title:  "Filter filenames"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.map-filenames"
    title:  "Map filenames"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.load-existing-meta"
    title:  "Load existing meta (if any)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.post-map-for-update"
    title:  "Post-map for update (with existing)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.compute-meta-diffs"
    title:  "Compute meta diffs"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.detect-orphans"
    title:  "Detect orphan meta files"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diff.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.validate.flow"
    title:  "Validate meta files only"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.validate.find-meta-files"
    title:  "Find *.thoth.yaml files"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.validate.parse-validate-yaml"
    title:  "Parse and validate YAML records"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.validate.collect-validation-results"
    title:  "Collect validation results"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.validate.write-json-result"
    title:  "Write JSON result (array/value/lines)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.parse-args"
    title:  "Parse args for diagnose"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.single-stage"
    title:  "Diagnose single stage"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.parse-subcommand-args"
    title:  "Parse args for diagnose"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.load-action-config"
    title:  "Load action config (CUE)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.resolve-target-step"
    title:  "Resolve target step"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.resolve-input-mode"
    title:  "Resolve input mode"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.dump-stage-input"
    title:  "Dump stage input (optional)"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.emit-run-header"
    title:  "Emit run header"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.execute-target-stage"
    title:  "Execute target stage"
    labels: ["call", "design", "flow"]
  },
  {
    name:   "call.diagnose.dump-stage-output"
    title:  "Dump stage output (optional)"
    labels: ["call", "design", "flow"]
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
