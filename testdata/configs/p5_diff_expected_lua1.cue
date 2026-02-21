{
  configVersion: "1"
  action: "diff-meta"
  discovery: { root: "testdata/repos/diff2" }
  diffMeta: {
    expectedLua: {
      inline: """
return function(locator, existingMeta)
  if locator == "a.txt" then
    return {
      a = 1,
      obj = { y = 2, z = 3 },
      arr = { 1, 2, 3 },
    }
  end
  return {
    only = true,
    extra = { k = 1 },
  }
end
"""
    }
  }
}
