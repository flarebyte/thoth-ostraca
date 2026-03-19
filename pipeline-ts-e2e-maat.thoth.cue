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
      return string.sub(locator, 1, 11) == "script/e2e/"
        and (
          string.sub(locator, -8) == ".test.ts"
          or string.sub(locator, -9) == ".suite.ts"
        )
      """
  }

  map: {
    inline: """
      return {
        language = "typescript",
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
      "--rules 'import_files_list,testcase_titles_list' " +
      "--language typescript " +
      "--json",
    ]
  }

  postMap: {
    inline: """
      local rules = shell and shell.json and shell.json.rules or {}
      local language = shell and shell.json and shell.json.language or "typescript"
      return {
        meta = {
          language = language,
          import_files_list = rules.import_files_list or {},
          testcase_titles_list = rules.testcase_titles_list or {},
        },
      }
      """
  }

  persistMeta: {
    enabled: true
    outDir: "./thoth-meta/ts-e2e"
  }

  output: {
    out: "./temp/pipeline-ts-e2e-maat.json"
    pretty: true
    lines: false
  }

  ui: {
    progress: true
    progressIntervalMs: 250
  }
}
