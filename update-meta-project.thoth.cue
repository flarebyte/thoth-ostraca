{
  configVersion: "1"
  action: "update-meta"

  discovery: {
    root: "internal"
    noGitignore: false
  }

  updateMeta: {
    patch: {
      language: "go"
    }
  }

  output: {
    out: "-"
    pretty: true
    lines: false
  }
}
