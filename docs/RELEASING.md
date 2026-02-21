# Releasing `thoth`

## Prerequisites
- `go`, `bun`, `gh` installed
- authenticated GitHub CLI session (`gh auth status`)

## Release command
```bash
make release
```

What it does:
- runs `release-check` (`lint`, `test`, `contract-snapshots`)
- builds multi-platform binaries into `build/` via `build-go.ts`
- creates `build/checksums.txt`
- creates GitHub release tag `v<version>` and uploads `build/*`

Version source:
- `main.project.yaml` -> `tags.version`

Dry run:
```bash
bun run release-go.ts --dry-run
```

## Distribution plan
- GitHub releases are the source artifacts.
- Homebrew distribution should be published from those artifacts through:
  - `https://github.com/flarebyte/homebrew-tap`
- Keep tap updates in sync with release tag and `checksums.txt`.
