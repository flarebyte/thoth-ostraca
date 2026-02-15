{
  configVersion: "v0"
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
    program: "this-program-does-not-exist-xyz"
    argsTemplate: ["-c", "echo '{json}'"]
    timeoutMs: 2000
  }
}

