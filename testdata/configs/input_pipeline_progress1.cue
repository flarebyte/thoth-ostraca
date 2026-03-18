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
        kind = "go",
      }
      """
  }
  shell: {
    enabled: true
    program: "sh"
    argsTemplate: [
      "-c",
      "printf '%s\\n' '{locator}'",
    ]
  }
  postMap: {
    inline: """
      return {
        meta = {
          kind = mapped and mapped.kind,
          shellLocator = shell and shell.stdout,
        },
      }
      """
  }
  persistMeta: {
    enabled: true
    dryRun: true
  }
  ui: {
    progress: true
    progressIntervalMs: 1
  }
}
