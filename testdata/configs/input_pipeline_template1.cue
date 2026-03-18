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
        slug = string.gsub(locator, "/", "__"),
      }
      """
  }
  shell: {
    enabled: true
    program: "sh"
    argsTemplate: [
      "-c",
      "printf '%s|%s|%s|%s|%s\\n' " +
      "'{locator}' '{file.base}' '{file.stem}' " +
      "'{file.ext}' '{mapped.slug}'",
    ]
  }
  postMap: {
    inline: """
      return {
        locator = locator,
        stdout = shell and shell.stdout,
      }
      """
  }
}
