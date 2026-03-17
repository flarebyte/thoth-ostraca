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
