{
  configVersion: "1"
  action: "input-pipeline"
  discovery: {
    root: "testdata/repos/input_pipeline1"
  }
  filter: {
    inline: """
      return string.sub(locator, -3) == ".go"
        and string.sub(locator, -8) ~= "_test.go"
      """
  }
  map: {
    inline: """
      return {
        locator = locator,
        kind = "go",
      }
      """
  }
  shell: {
    enabled: true
    program: "sh"
    argsTemplate: [
      "-c",
      "printf '%s\\n' '{json}'",
    ]
  }
  postMap: {
    inline: """
      return {
        locator = locator,
        kind = mapped and mapped.kind,
        exit = shell and shell.exitCode,
      }
      """
  }
  reduce: {
    inline: """
      return (acc or 0) + 1
      """
  }
}
