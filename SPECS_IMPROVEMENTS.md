# Specs Improvements

## Must-Have

- `create-meta` must support writing generated `.thoth.yaml` files to a dedicated output directory instead of only alongside source files.
- The output directory must preserve source-relative paths so generated metadata stays deterministic and collision-free.
- The feature must support a dry-run mode that reports planned writes without modifying the repository.
- Input-file discovery must exclude VCS and tool-internal directories such as `.git` by default.
