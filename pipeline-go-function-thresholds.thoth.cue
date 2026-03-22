{
  configVersion: "1"
  action: "input-pipeline"

  discovery: {
    root: "."
  }

  errors: {
    mode: "keep-going"
    embedErrors: true
  }

  filter: {
    inline: """
      return thoth.starts_with(locator, "internal/metafile/")
        and thoth.ends_with(locator, ".go")
        and not thoth.ends_with(locator, "_test.go")
        and locator == "internal/metafile/write.go"
      """
  }

  map: {
    inline: """
      return {
        language = "go",
        tokenThreshold = 100,
        complexityThreshold = 4,
      }
      """
  }

  shell: {
    enabled: true
    decodeJsonStdout: true
    program: "/bin/sh"
    workingDir: "."
    argsTemplate: [
      "-c",
      "npx maat-ostraca analyse " +
      "--in '{locator}' " +
      "--rules 'function_map' " +
      "--language go " +
      "--json",
    ]
  }

  postMap: {
    inline: """
      local function worseFirst(a, b)
        if a.cognitiveComplexity ~= b.cognitiveComplexity then
          return a.cognitiveComplexity > b.cognitiveComplexity
        end
        if a.tokens ~= b.tokens then
          return a.tokens > b.tokens
        end
        return a.name < b.name
      end

      local rules = shell and shell.json and shell.json.rules or {}
      local functionMap = rules.function_map or {}
      local flagged = {}
      local tokenThreshold = mapped and mapped.tokenThreshold or 100
      local complexityThreshold = mapped and mapped.complexityThreshold or 4

      for _, name in ipairs(thoth.sort_keys(functionMap)) do
        local node = functionMap[name] or {}
        local tokens = node.tokens or 0
        local complexity = node.cognitiveComplexity or 0
        if tokens > tokenThreshold or complexity > complexityThreshold then
          thoth.push(flagged, {
            name = name,
            cognitiveComplexity = node.cognitiveComplexity or 0,
            cyclomaticComplexity = node.cyclomaticComplexity or 0,
            tokens = node.tokens or 0,
            loc = node.loc or 0,
            sloc = node.sloc or 0,
          })
        end
      end

      table.sort(flagged, worseFirst)

      return {
        locator = locator,
        flaggedCount = #flagged,
        flaggedFunctions = flagged,
      }
      """
  }

  output: {
    out: "./temp/pipeline-go-function-thresholds.json"
    pretty: true
    lines: false
  }

  ui: {
    progress: true
    progressIntervalMs: 250
  }
}
