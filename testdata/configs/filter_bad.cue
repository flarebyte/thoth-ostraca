{
  configVersion: "1"
  action: "nop"
  discovery: {
    root: "testdata/repos/yaml1"
  }
  filter: {
    inline: "return (meta and meta.enabled) =="  // syntax error
  }
}

