{
  configVersion: "1"
  action: "search"

  discovery: {
    root: "."
    noGitignore: false
  }

  filter: {
    inline: "return string.sub(locator or \"\", -3) == \".go\""
  }

  map: {
    inline: "return { locator = locator }"
  }

  output: {
    out: "-"
    pretty: true
    lines: false
  }
}
