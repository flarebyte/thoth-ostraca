{
  configVersion: "v0"
  action: "diff-meta"
  discovery: { root: "testdata/repos/diff2" }
  diffMeta: {
    expectedPatch: {
      a: 1
      obj: { y: 9, z: 3 }
      arr: [1, 2, 3]
    }
  }
}
