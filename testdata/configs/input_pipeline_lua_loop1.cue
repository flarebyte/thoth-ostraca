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
      "printf '%s\\n' '{locator}'",
    ]
  }
  postMap: {
    inline: """
      local raw = shell and shell.stdout or ""
      local clean = string.gsub(raw, "\\n", "")
      local chars = {}
      for i = 1, string.len(clean) do
        chars[i] = string.sub(clean, i, i)
      end
      return {
        locator = locator,
        count = #chars,
        chars = chars,
      }
      """
  }
}
