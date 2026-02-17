{
  configVersion: "v0"
  action: "pipeline"
  discovery: {
    root: "testdata/repos/p3_locator_urls1"
  }
  locatorPolicy: {
    allowURLs: true
  }
  errors: {
    mode: "keep-going"
    embedErrors: true
  }
}
