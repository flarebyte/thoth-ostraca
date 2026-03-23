// File Guide for dev/ai agents:
// Purpose: Build the structured Lua-visible record context shared by filter, map, postMap, and reduce scripts.
// Responsibilities:
// - Expose locator, meta, fileInfo, and git fields in a consistent object shape.
// - Normalize nested fileInfo and git values into plain Lua-friendly maps.
// - Omit unavailable optional sections to keep the script contract compact.
// Architecture notes:
// - This context intentionally mirrors only stable record fields; transient shell or postMap state is provided separately by the calling stage.
// - Git and fileInfo are converted to plain maps here so each Lua stage does not need to duplicate shape-conversion logic.
package stage

func luaRecordContext(rec Record) map[string]any {
	ctx := map[string]any{
		"locator": rec.Locator,
		"meta":    rec.Meta,
	}
	if rec.FileInfo != nil {
		ctx["fileInfo"] = map[string]any{
			"size":    rec.FileInfo.Size,
			"mode":    rec.FileInfo.Mode,
			"modTime": rec.FileInfo.ModTime,
			"isDir":   rec.FileInfo.IsDir,
		}
	}
	if rec.Git != nil {
		git := map[string]any{
			"tracked": rec.Git.Tracked,
			"ignored": rec.Git.Ignored,
			"status":  rec.Git.Status,
		}
		if rec.Git.LastCommit != nil {
			git["lastCommit"] = map[string]any{
				"hash":   rec.Git.LastCommit.Hash,
				"author": rec.Git.LastCommit.Author,
				"time":   rec.Git.LastCommit.Time,
			}
		}
		ctx["git"] = git
	}
	return ctx
}
