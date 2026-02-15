package stage

import (
	"context"
	"fmt"

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
	applyMinimalToMeta(&out, min)
	// Allowed actions (Phase 2): pipeline, validate, create-meta (accept "nop" as alias for pipeline for existing tests)
	if min.Action != "pipeline" && min.Action != "validate" && min.Action != "create-meta" && min.Action != "nop" {
		return Envelope{}, fmt.Errorf("invalid action: allowed 'pipeline' or 'validate'")
	}
	return out, nil
}

type ErrMissingConfigPath struct{}

func (ErrMissingConfigPath) Error() string { return "missing required meta.configPath" }

func init() { Register("validate-config", ValidateConfig) }
