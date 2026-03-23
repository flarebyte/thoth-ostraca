# Lua Helpers

`thoth` exposes a small helper table in CUE-embedded Lua:

- `thoth.all(list, predicate)`
- `thoth.any(list, predicate)`
- `thoth.contains(list, value)`
- `thoth.copy(tbl)`
- `thoth.deep_copy(tbl)`
- `thoth.ends_with(s, suffix)`
- `thoth.filter(list, fn)`
- `thoth.find(list, predicate)`
- `thoth.flatten(list)`
- `thoth.is_empty(tbl)`
- `thoth.map(list, fn)`
- `thoth.push(list, value)`
- `thoth.reduce(list, init, fn)`
- `thoth.sort_keys(tbl)`
- `thoth.sort_values(tbl)`
- `thoth.starts_with(s, prefix)`
- `thoth.split(s, sep)`
- `thoth.trim(s)`

Only the sandboxed Lua surface is available. There is no `require(...)`,
filesystem access, or network access from Lua.

## Examples

Filter Go files:

```lua
return thoth.ends_with(locator, ".go")
  and not thoth.ends_with(locator, "_test.go")
```

Filter a subtree:

```lua
return thoth.starts_with(locator, "internal/")
```

Get deterministic object keys:

```lua
local functionMap = rules.function_map or {}
local names = thoth.sort_keys(functionMap)
```

Filter and map a list:

```lua
local goFiles = thoth.filter(files, function(file)
  return thoth.ends_with(file, ".go")
end)

local trimmed = thoth.map(lines, function(line)
  return thoth.trim(line)
end)
```

Reduce a list:

```lua
local total = thoth.reduce(numbers, 0, function(acc, item)
  return acc + item
end)
```

Flatten nested arrays:

```lua
local flat = thoth.flatten({
  "a",
  {"b", {"c"}},
})
```

## Notes

- `thoth.sort_keys(tbl)` sorts string keys only.
- `thoth.sort_values(tbl)` sorts string values only.
- `thoth.copy(tbl)` is shallow.
- `thoth.deep_copy(tbl)` recursively copies nested tables.
- `thoth.push(list, value)` mutates the provided list and also returns it.
- Higher-order helpers such as `map`, `filter`, `find`, `any`, `all`, and
  `reduce` call your Lua callback inside the sandbox.
