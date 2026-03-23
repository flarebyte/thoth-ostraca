{
  configVersion: "1"
  action: "nop"

  discovery: {
    root: "thoth-meta"
  }

  map: {
    inline: """
      return {
        locator = locator,
        language = meta and meta.language,
        purpose = meta and meta.purpose,
        responsibilities = meta and meta.responsibilities,
        architecture_notes = meta and meta.architecture_notes,
      }
      """
  }

  postMap: {
    inline: """
      return mapped
      """
  }

  reduce: {
    inline: """
      local acc0 = acc or {
        count = 0,
        files = {},
      }

      acc0.count = acc0.count + 1
      thoth.push(acc0.files, item)
      return acc0
      """
  }

  output: {
    out: "temp/thoth-meta-aggregate.json"
    pretty: true
    lines: false
  }
}
