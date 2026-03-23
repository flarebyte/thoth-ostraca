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
  shell: {
    enabled: true
    program: "sh"
    argsTemplate: [
      "-c",
      "printf '%s\\n' '{unknown}'",
    ]
  }
  postMap: {
    inline: """
      return {
        locator = locator,
      }
      """
  }
}
