{
  configVersion: "1"
  action: "input-pipeline"
  discovery: {
    root: "testdata/repos/input_pipeline1"
  }
  filter: {
    inline: """
      return locator == "a.go"
      """
  }
  map: {
    inline: """
      return {
        locator = locator,
      }
      """
  }
  shell: {
    enabled: true
    decodeJsonStdout: true
    program: "sh"
    argsTemplate: [
      "-c",
      "printf 'not-json\\n'",
    ]
  }
}
