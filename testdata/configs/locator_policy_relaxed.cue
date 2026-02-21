{
  configVersion: "1"
  action: "nop"
  discovery: {
    root: "testdata/repos/locator1"
  }
  errors: {
    mode: "keep-going"
    embedErrors: true
  }
  locatorPolicy: {
    allowParentRefs: true
    posixStyle: false
  }
}

