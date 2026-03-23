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
    program: "/bin/sh"
    workingDir: "does-not-exist"
    argsTemplate: ["-c", "printf '%s\\n' '{json}'"]
    timeoutMs: 2000
  }
}
