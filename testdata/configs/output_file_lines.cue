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
  output: {
    out: "temp/out.lines.ndjson"
    pretty: false
    lines: true
  }
}
