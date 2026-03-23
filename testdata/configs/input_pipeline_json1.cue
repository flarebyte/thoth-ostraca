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
    decodeJsonStdout: true
    program: "sh"
    argsTemplate: [
      "-c",
      "printf '%s\\n' '{json}'",
    ]
  }
  postMap: {
    inline: """
      return {
        locator = shell and shell.json and shell.json.locator,
        kind = shell and shell.json and shell.json.kind,
        exit = shell and shell.exitCode,
      }
      """
  }
}
