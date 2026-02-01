# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
Parse CLI args for metadata find
  Find individual meta locators
    Filter metadata for a locator using a Lua script
```

Supported use cases:

  - Filter metadata associated with locator

Unsupported use cases (yet):

  - Map metadata associated with locator
  - Reduce metadata for all locators
  - Load map reduce action config from file (YAML)
  - Filter map and reduce can be scripted (Lua)
  - Use map metadata for running shell with locator name
  - Locator can be (relative) file or url
  - Processing is done in parallel
  - batch creation of locators metadata
  - batch update of locators metadata
  - batch diff of locators metadata
