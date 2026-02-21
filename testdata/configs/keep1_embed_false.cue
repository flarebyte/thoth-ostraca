{
  configVersion: "1"
  action: "nop"
  discovery: {
    root: "testdata/repos/keep1"
  }
  filter: {
    inline: "return true"
  }
  map: {
    inline: "if (meta and meta.name) == \"LuaErr\" then error(\"boom\") end; return { locator = locator, name = meta and meta.name }"
  }
  errors: {
    mode: "keep-going"
    embedErrors: false
  }
}

