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
      return string.sub(locator, -8) == "_test.go"
      """
  }

  map: {
    inline: """
      return {
        language = "go",
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
      "--rules 'import_files_list,function_map' " +
      "--language go " +
      "--json",
    ]
  }

  postMap: {
    inline: """
      local function getKeysSorted(t)
        local keys = {}
        for key, _ in pairs(t) do
          table.insert(keys, key)
        end
        table.sort(keys)
        return keys
      end
      local rules = shell and shell.json and shell.json.rules or {}
      return {
        meta = {
          language = mapped and mapped.language or "go",
          import_files_list = rules.import_files_list or {},
          function_list = getKeysSorted(rules.function_map or  {}),
        },
      }
      """
  }

  persistMeta: {
    enabled: true
    outDir: "./thoth-meta/go-test"
  }

  output: {
    out: "./temp/pipeline-go-test-maat.json"
    pretty: true
    lines: false
  }

  ui: {
    progress: true
    progressIntervalMs: 250
  }
}
