{
  configVersion: "v0"
  action: "pipeline"
  discovery: {
    root: "testdata/repos/p3_yaml_strict1"
  }
  validation: {
    allowUnknownTopLevel: true
  }
  errors: {
    mode: "keep-going"
    embedErrors: true
  }
}
