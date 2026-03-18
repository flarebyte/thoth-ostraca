{
  configVersion: "1"
  action: "input-pipeline"

  discovery: {
    root: "internal"
  }

  errors: {
    mode: "keep-going"
    embedErrors: true
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
        language = "go",
      }
      """
  }

  shell: {
    enabled: true
    decodeJsonStdout: true
    program: "sh"
    workingDir: "internal"
    argsTemplate: [
      "-c",
      "npx maat-ostraca analyse " +
      "--in '{locator}' " +
      "--rules 'import_files_list,package_imports_list' " +
      "--language go " +
      "--json",
    ]
  }

  postMap: {
    inline: """
      local rules = shell and shell.json and shell.json.rules or {}
      return {
        meta = {
          language = mapped and mapped.language or "go",
          import_files_list = rules.import_files_list or {},
          package_imports_list = rules.package_imports_list or {},
        },
      }
      """
  }

  persistMeta: {
    enabled: true
    outDir: "./temp/pipeline-go-maat-sidecars"
  }

  output: {
    out: "./temp/pipeline-go-maat.json"
    pretty: true
    lines: false
  }

  ui: {
    progress: true
    progressIntervalMs: 250
  }
}
