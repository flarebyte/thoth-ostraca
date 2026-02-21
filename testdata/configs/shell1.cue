{
  configVersion: "1"
  action: "nop"
  discovery: {
    root: "testdata/repos/yaml1"
  }
  filter: {
    inline: "return (meta and meta.enabled) == true"
  }
  map: {
    inline: "return { locator = locator, name = meta and meta.name }"
  }
  shell: {
    enabled: true
    program: "sh"
    argsTemplate: ["-c", "echo '{json}'"]
  }
  postMap: {
    inline: "return { locator = locator, exit = (shell and shell.exitCode) }"
  }
}

