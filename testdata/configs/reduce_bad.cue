{
  configVersion: "v0"
  action: "nop"
  discovery: {
    root: "testdata/repos/yaml1"
  }
  reduce: {
    inline: "return (acc or 0) +"  // syntax error
  }
}

