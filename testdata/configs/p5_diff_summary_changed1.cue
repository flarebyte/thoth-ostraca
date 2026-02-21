{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "testdata/repos/p5_diff_arrays1" }
  diffMeta: {
    format: "summary"
    only: "changed"
    summary: true
    expectedPatch: {
      kind: []
      rules: [
        { tags: ["a", "c"] },
        { tags: ["x", "y"] },
        { tags: [] },
      ]
      value: "1"
    }
  }
}
