{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "testdata/repos/p5_diff_arrays1" }
  diffMeta: {
    expectedPatch: {
      value: "1"
      kind: []
      rules: [
        { tags: ["a", "c"] },
        { tags: ["x", "y"] },
        { tags: [] },
      ]
    }
  }
}
