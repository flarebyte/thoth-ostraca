// File Guide for dev/ai agents:
// Purpose: Copy validated discovery, validation, and limits config fields into the runtime envelope metadata model.
// Responsibilities:
// - Apply discovery root and include/exclude settings when present in the parsed config.
// - Apply validation flags such as allowUnknownTopLevel.
// - Apply memory and YAML size limits with the runtime defaults preserved elsewhere.
// Architecture notes:
// - Config application is split by concern so schema growth does not force one monolithic applyMinimalToMeta function.
// - These helpers only copy parsed values; validation of those values belongs in validate_config_common_validation.go.
package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

func applyDiscoveryMeta(out *Envelope, min config.Minimal) {
	if min.Discovery.HasRoot ||
		min.Discovery.HasInclude ||
		min.Discovery.HasExclude ||
		min.Discovery.HasNoGitignore ||
		min.Discovery.HasFollowSymlink {
		if out.Meta.Discovery == nil {
			out.Meta.Discovery = &DiscoveryMeta{}
		}
		if min.Discovery.HasRoot {
			out.Meta.Discovery.Root = min.Discovery.Root
		}
		if min.Discovery.HasInclude {
			out.Meta.Discovery.Include = append(
				[]string(nil),
				min.Discovery.Include...,
			)
		}
		if min.Discovery.HasExclude {
			out.Meta.Discovery.Exclude = append(
				[]string(nil),
				min.Discovery.Exclude...,
			)
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
