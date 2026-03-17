{
  configVersion: "1"
  action: "pipeline"

  discovery: {
    root: "internal"
    noGitignore: false
  }

  errors: {
    mode: "keep-going"
    embedErrors: true
  }

  lua: {
    instructionLimit: 3000000
  }

  filter: {
    inline:
      "return string.sub(locator or \"\", -3) == \".go\" " +
      "and string.sub(locator or \"\", -8) ~= \"_test.go\""
  }

  map: {
    inline: "return locator"
  }

  shell: {
    enabled: true
    program: "/bin/sh"
    workingDir: "internal"
    strictTemplating: false
    argsTemplate: [
      "-c",
      "path={json}; " +
        "path=${path#\\\"}; " +
        "path=${path%\\\"}; " +
        "npx maat-ostraca analyse " +
        "--in \"$path\" " +
        "--rules 'import_files_list,package_imports_list' " +
        "--language go " +
        "--json",
    ]
  }

  postMap: {
    inline:
      "local function pick(json, key) " +
      "local body = string.match(" +
      "json or \"\", " +
      "\"\\\"\" .. key .. \"\\\":%[(.-)%]\") " +
      "local out = {} " +
      "if not body then return out end " +
      "for item in string.gmatch(body, \"\\\"([^\\\"]+)\\\"\") do " +
      "out[#out + 1] = item " +
      "end " +
      "return out " +
      "end; " +
      "local raw = shell and shell.stdout or \"\"; " +
      "return { " +
      "locator = locator, " +
      "nextMeta = { " +
      "language = (meta and meta.language) or \"go\", " +
      "import_files_list = pick(raw, \"import_files_list\"), " +
      "package_imports_list = pick(raw, \"package_imports_list\") " +
      "} " +
      "}"
  }

  output: {
    out: "./temp/pipeline-go-maat.json"
    pretty: true
    lines: false
  }
}
