{
  configVersion: "1"
  action: "input-pipeline"
  discovery: {
    root: "testdata/repos/input_discovery_safe1"
  }
  filter: {
    inline: "return true"
  }
  map: {
    inline: "return { locator = locator }"
  }
  postMap: {
    inline: "return { locator = locator }"
  }
}
