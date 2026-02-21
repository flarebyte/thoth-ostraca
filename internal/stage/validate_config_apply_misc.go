package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

func applyUpdateMeta(out *Envelope, min config.Minimal) {
	if !min.UpdateMeta.HasPatch && !min.UpdateMeta.HasExpectedLuaCode {
		return
	}
	if out.Meta.UpdateMeta == nil {
		out.Meta.UpdateMeta = &UpdateMetaMeta{}
	}
	if min.UpdateMeta.HasPatch {
		if cp, ok := deepCopyAny(min.UpdateMeta.Patch).(map[string]any); ok {
			out.Meta.UpdateMeta.Patch = cp
		} else {
			out.Meta.UpdateMeta.Patch = map[string]any{}
		}
	}
	if min.UpdateMeta.HasExpectedLuaCode {
		out.Meta.UpdateMeta.ExpectedLuaInline = min.UpdateMeta.ExpectedLuaInline
	}
}

func applyDiffMeta(out *Envelope, min config.Minimal) {
	if min.Action != "diff-meta" {
		return
	}
	if out.Meta.DiffMeta == nil {
		out.Meta.DiffMeta = &DiffMetaMeta{ExpectedPatch: map[string]any{}, Format: "summary", FailOnChange: false}
	}
	out.Meta.DiffMeta.Format = "summary"
	out.Meta.DiffMeta.FailOnChange = false
	out.Meta.DiffMeta.ExpectedLuaInline = ""
	if min.DiffMeta.HasFormat {
		out.Meta.DiffMeta.Format = min.DiffMeta.Format
	}
	if min.DiffMeta.HasFailOnChange {
		out.Meta.DiffMeta.FailOnChange = min.DiffMeta.FailOnChange
	}
	if min.DiffMeta.HasExpectedLuaCode {
		out.Meta.DiffMeta.ExpectedLuaInline = min.DiffMeta.ExpectedLuaInline
	}
	if min.DiffMeta.HasExpectedPatch {
		if cp, ok := deepCopyAny(min.DiffMeta.ExpectedPatch).(map[string]any); ok {
			out.Meta.DiffMeta.ExpectedPatch = cp
		} else {
			out.Meta.DiffMeta.ExpectedPatch = map[string]any{}
		}
	}
}

func applyErrorsMeta(out *Envelope, min config.Minimal) {
	if min.Errors.HasMode || min.Errors.HasEmbed {
		if out.Meta.Errors == nil {
			out.Meta.Errors = &ErrorsMeta{}
		}
		if min.Errors.HasMode {
			out.Meta.Errors.Mode = min.Errors.Mode
		}
		if min.Errors.HasEmbed {
			out.Meta.Errors.EmbedErrors = min.Errors.EmbedErrors
		}
	}
}

func applyFileInfoMeta(out *Envelope, min config.Minimal) {
	if !min.FileInfo.HasEnabled {
		return
	}
	if out.Meta.FileInfo == nil {
		out.Meta.FileInfo = &FileInfoMeta{}
	}
	out.Meta.FileInfo.Enabled = min.FileInfo.Enabled
}

func applyGitMeta(out *Envelope, min config.Minimal) {
	if !min.Git.HasEnabled {
		return
	}
	if out.Meta.Git == nil {
		out.Meta.Git = &GitMeta{}
	}
	out.Meta.Git.Enabled = min.Git.Enabled
}

func applyWorkersMeta(out *Envelope, min config.Minimal) {
	if min.Workers.HasCount {
		out.Meta.Workers = min.Workers.Count
	}
}

func applyUIMeta(out *Envelope, min config.Minimal) {
	if !min.UI.HasSection {
		return
	}
	if out.Meta.UI == nil {
		out.Meta.UI = &UIMeta{Progress: false, ProgressIntervalMs: defaultUIProgressIntervalMs}
	}
	if min.UI.HasProgress {
		out.Meta.UI.Progress = min.UI.Progress
	}
	if min.UI.HasIntervalMs {
		out.Meta.UI.ProgressIntervalMs = min.UI.ProgressIntervalMs
	}
	if out.Meta.UI.ProgressIntervalMs <= 0 {
		out.Meta.UI.ProgressIntervalMs = defaultUIProgressIntervalMs
	}
}

func applyLocatorPolicyMeta(out *Envelope, min config.Minimal) {
	if (min.LocatorPolicy.HasAllowAbs || min.LocatorPolicy.HasAllowParent || min.LocatorPolicy.HasPosix || min.LocatorPolicy.HasAllowURLs) || out.Meta.LocatorPolicy != nil {
		if out.Meta.LocatorPolicy == nil {
			out.Meta.LocatorPolicy = &LocatorPolicy{
				AllowAbsolute:   false,
				AllowParentRefs: false,
				PosixStyle:      true,
				AllowURLs:       false,
			}
		}
		if min.LocatorPolicy.HasAllowAbs {
			out.Meta.LocatorPolicy.AllowAbsolute = min.LocatorPolicy.AllowAbsolute
		}
		if min.LocatorPolicy.HasAllowParent {
			out.Meta.LocatorPolicy.AllowParentRefs = min.LocatorPolicy.AllowParentRefs
		}
		if min.LocatorPolicy.HasPosix {
			out.Meta.LocatorPolicy.PosixStyle = min.LocatorPolicy.PosixStyle
		}
		if min.LocatorPolicy.HasAllowURLs {
			out.Meta.LocatorPolicy.AllowURLs = min.LocatorPolicy.AllowURLs
		}
	}
}
