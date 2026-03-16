package flyb

source: "docs-designs-flow-design-md"
name:   "thoth-design-docs"

modules: [
  "design",
  "documentation",
  "thoth",
]

reports: [{
  title:       "Flow Design Overview"
  filepath:    "../design/FLOW_DESIGN.md"
  description: "Migrated from docs/designs/FLOW_DESIGN.md."
  sections: [
    {
      title: "Function calls tree"
      sections: [
        {
          title: "Flow call graph"
          arguments: [
            "graph-subject-label=call",
            "graph-edge-label=delegate_to",
            "graph-start-node=call.thoth.cli.root",
            "graph-renderer=markdown-text",
            "cycle-policy=disallow",
          ]
          notes: ["call.thoth.cli.root"]
        },
        {
          title: "Supported use cases"
          notes: #functionCallsTreeUseCaseRefs
        },
      ]
    },
    {
      title: "Suggested Go Implementation"
      sections: [{
        title: "Details"
        notes: ["flow.suggested-go-implementation"]
      }]
    },
    {
      title: "Exit Codes"
      sections: [{
        title: "Details"
        notes: ["flow.exit.codes"]
      }]
    },
    {
      title: "Ordering & Determinism"
      sections: [{
        title: "Details"
        notes: ["flow.ordering-determinism"]
      }]
    },
    {
      title: "Action Config (CUE Example)"
      sections: [{
        title: "Details"
        notes: ["flow.action-config-cue-example"]
      }]
    },
    {
      title: "Action Config (Create Example, CUE)"
      sections: [{
        title: "Details"
        notes: ["flow.action-config-create-example-cue"]
      }]
    },
    {
      title: "Action Config (Create Minimal, CUE)"
      sections: [{
        title: "Details"
        notes: ["flow.action-config-create-minimal-cue"]
      }]
    },
    {
      title: "Action Config (Diff Example, CUE)"
      sections: [{
        title: "Details"
        notes: ["flow.action-config-diff-example-cue"]
      }]
    },
    {
      title: "Action Config (Lua Limits, CUE)"
      sections: [{
        title: "Details"
        notes: ["flow.action-config-lua-limits-cue"]
      }]
    },
    {
      title: "Lua Data Contracts"
      sections: [{
        title: "Details"
        notes: ["flow.lua-data-contracts"]
      }]
    },
    {
      title: "Diagnose Stage Boundary Types (Examples)"
      sections: [{
        title: "Details"
        notes: ["flow.diagnose-stage-boundary-types-examples"]
      }]
    },
    {
      title: "Lua Input Examples"
      sections: [{
        title: "Details"
        notes: ["flow.lua-input-examples"]
      }]
    },
    {
      title: "Reduce Behavior"
      sections: [{
        title: "Details"
        notes: ["flow.reduce-behavior"]
      }]
    },
    {
      title: "Error Handling"
      sections: [{
        title: "Details"
        notes: ["flow.error-handling"]
      }]
    },
    {
      title: "Result Shapes"
      sections: [{
        title: "Details"
        notes: ["flow.result-shapes"]
      }]
    },
    {
      title: "Diff Output Shape"
      sections: [{
        title: "Details"
        notes: ["flow.diff-output-shape"]
      }]
    },
    {
      title: "Update Merge Strategy"
      sections: [{
        title: "Details"
        notes: ["flow.update-merge-strategy"]
      }]
    },
    {
      title: "Lua Builtins (thoth namespace)"
      sections: [{
        title: "Details"
        notes: ["flow.lua-builtins-thoth-namespace"]
      }]
    },
    {
      title: "Locator Normalization"
      sections: [{
        title: "Details"
        notes: ["flow.locator-normalization"]
      }]
    },
    {
      title: "Discovery Semantics"
      sections: [{
        title: "Details"
        notes: ["flow.discovery-semantics"]
      }]
    },
    {
      title: "Lua Execution Environment"
      sections: [{
        title: "Details"
        notes: ["flow.lua-execution-environment"]
      }]
    },
    {
      title: "Shell Execution Spec"
      sections: [{
        title: "Details"
        notes: ["flow.shell-execution-spec"]
      }]
    },
    {
      title: "Function calls details"
      sections: [{
        title: "Details"
        notes: #callDetailRefs
      }]
    },
    {
      title: "Action Script Scope"
      sections: [{
        title: "Details"
        notes: ["flow.action-script-scope"]
      }]
    },
    {
      title: "Pure helper functions"
      sections: [{
        title: "Details"
        notes: ["flow.pure-helper-functions"]
      }]
    },
    {
      title: "Go Package Outline"
      sections: [{
        title: "Details"
        notes: ["flow.go-package-outline"]
      }]
    },
    {
      title: "Design Decisions"
      sections: [{
        title: "Details"
        notes: ["flow.design-decisions"]
      }]
    },
    {
      title: "Diagnose Command"
      sections: [{
        title: "Details"
        notes: ["flow.diagnose-command"]
      }]
    },
    {
      title: "Filename Collision & Stability"
      sections: [{
        title: "Details"
        notes: ["flow.filename-collision-stability"]
      }]
    },
    {
      title: "Schema Validation"
      sections: [{
        title: "Details"
        notes: ["flow.schema-validation"]
      }]
    },
    {
      title: "Config Schema & Versioning"
      sections: [{
        title: "Details"
        notes: ["flow.config-schema-versioning"]
      }]
    },
    {
      title: "CUE Tips (Inline Lua)"
      sections: [{
        title: "Details"
        notes: ["flow.cue-tips-inline-lua"]
      }]
    },
    {
      title: "Stage Contracts"
      sections: [{
        title: "Details"
        notes: ["flow.stage-contracts"]
      }]
    },
    {
      title: "Open Design Questions"
      sections: [{
        title: "Details"
        notes: ["flow.open-design-questions"]
      }]
    },
  ]
}]

argumentRegistry: {
  version: "1"
  arguments: []
}
