package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

func applyDiscoveryMeta(out *Envelope, min config.Minimal) {
	if min.Discovery.HasRoot || min.Discovery.HasNoGitignore || min.Discovery.HasFollowSymlink {
		if out.Meta.Discovery == nil {
			out.Meta.Discovery = &DiscoveryMeta{}
		}
		if min.Discovery.HasRoot {
			out.Meta.Discovery.Root = min.Discovery.Root
		}
		if min.Discovery.HasNoGitignore {
			out.Meta.Discovery.NoGitignore = min.Discovery.NoGitignore
		}
		if min.Discovery.HasFollowSymlink {
			out.Meta.Discovery.FollowSymlinks = min.Discovery.FollowSymlinks
		}
	}
}

func applyValidationMeta(out *Envelope, min config.Minimal) {
	if min.Validation.HasAllowUnknownTop {
		if out.Meta.Validation == nil {
			out.Meta.Validation = &ValidationMeta{}
		}
		out.Meta.Validation.AllowUnknownTopLevel = min.Validation.AllowUnknownTopLevel
	}
}

func applyLimitsMeta(out *Envelope, min config.Minimal) {
	if out.Meta.Limits == nil {
		out.Meta.Limits = &LimitsMeta{MaxRecordsInMemory: defaultMaxRecordsInMemory}
	}
	if min.Limits.HasMaxYAMLBytes {
		out.Meta.Limits.MaxYAMLBytes = min.Limits.MaxYAMLBytes
	}
	if min.Limits.HasMaxRecordsInMemory {
		out.Meta.Limits.MaxRecordsInMemory = min.Limits.MaxRecordsInMemory
	}
}
