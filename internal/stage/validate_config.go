package stage

import (
	"context"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

// ValidateConfig is the stage implementation for "validate-config".
func ValidateConfig(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Extract configPath from meta
	if in.Meta == nil || in.Meta.ConfigPath == "" {
		return Envelope{}, ErrMissingConfigPath{}
	}
	// Parse and validate minimal config
	min, err := config.ParseMinimal(in.Meta.ConfigPath)
	if err != nil {
		return Envelope{}, err
	}
	out := in
	if out.Meta == nil {
		out.Meta = &Meta{}
	}
	out.Meta.Config = &ConfigMeta{ConfigVersion: min.ConfigVersion, Action: min.Action}
	// Do not persist configPath in output
	out.Meta.ConfigPath = ""
	// Optionally expose discovery settings if present
	if min.Discovery.HasRoot || min.Discovery.HasNoGitignore {
		if out.Meta.Discovery == nil {
			out.Meta.Discovery = &DiscoveryMeta{}
		}
		if min.Discovery.HasRoot {
			out.Meta.Discovery.Root = min.Discovery.Root
		}
		if min.Discovery.HasNoGitignore {
			out.Meta.Discovery.NoGitignore = min.Discovery.NoGitignore
		}
	}
	// Optionally expose filter.inline
	if min.Filter.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.FilterInline = min.Filter.Inline
	}
	// Optionally expose map.inline
	if min.Map.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.MapInline = min.Map.Inline
	}
	return out, nil
}

type ErrMissingConfigPath struct{}

func (ErrMissingConfigPath) Error() string { return "missing required meta.configPath" }

func init() { Register("validate-config", ValidateConfig) }
