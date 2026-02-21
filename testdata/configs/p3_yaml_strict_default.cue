{
  configVersion: "1"
  action: "pipeline"
  discovery: {
    root: "testdata/repos/p3_yaml_strict1"
  }
  validation: {
    allowUnknownTopLevel: false
  }
  errors: {
    mode: "keep-going"
    embedErrors: true
  }
}
