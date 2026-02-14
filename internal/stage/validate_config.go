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
	return out, nil
}

type ErrMissingConfigPath struct{}

func (ErrMissingConfigPath) Error() string { return "missing required meta.configPath" }

func init() { Register("validate-config", ValidateConfig) }
