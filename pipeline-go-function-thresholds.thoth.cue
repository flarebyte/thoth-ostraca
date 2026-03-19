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
      return string.sub(locator, 1, 18) == "internal/metafile/"
        and string.sub(locator, -3) == ".go"
        and string.sub(locator, -8) ~= "_test.go"
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
      local function appendFlagged(flagged, name, node)
        flagged[#flagged + 1] = {
          name = name,
          cognitiveComplexity = node.cognitiveComplexity or 0,
          cyclomaticComplexity = node.cyclomaticComplexity or 0,
          tokens = node.tokens or 0,
          loc = node.loc or 0,
          sloc = node.sloc or 0,
        }
      end

      local function sortByName(items)
        table.sort(items, function(a, b)
          return a.name < b.name
        end)
      end

      local rules = shell and shell.json and shell.json.rules or {}
      local functionMap = rules.function_map or {}
      local flagged = {}
      local tokenThreshold = mapped and mapped.tokenThreshold or 100
      local complexityThreshold = mapped and mapped.complexityThreshold or 4

      for name, node in pairs(functionMap) do
        local tokens = node.tokens or 0
        local complexity = node.cognitiveComplexity or 0
        if tokens > tokenThreshold or complexity > complexityThreshold then
          appendFlagged(flagged, name, node)
        end
      end

      sortByName(flagged)

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
