{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "testdata/repos/diff2" }
  diffMeta: {
    only: "unchanged"
    expectedPatch: {
      a: "1"
      arr: [1, 2, 3]
      obj: { y: 9, z: 3 }
    }
  }
}
