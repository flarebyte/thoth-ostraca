// File Guide for dev/ai agents:
// Purpose: Parse and validate the requested config file early so the run pipeline works from one normalized minimal config contract.
// Responsibilities:
// - Require a config path in the input envelope.
// - Parse the minimal config shape and apply shared validation rules.
// - Copy validated config fields into envelope metadata and reject unsupported actions.
// Architecture notes:
// - This stage validates only the minimal contract needed to build the stage graph; richer action-specific checks stay in shared validation helpers.
// - The explicit action allowlist is intentional so newly added actions are opt-in and cannot silently bypass stage wiring review.
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
	if err := validateCommonConfig(min); err != nil {
		return Envelope{}, err
	}
	out := in
	applyMinimalToMeta(&out, min)
	if min.Action != "pipeline" &&
		min.Action != "input-pipeline" &&
		min.Action != "validate" &&
		min.Action != "create-meta" &&
		min.Action != "update-meta" &&
		min.Action != "diff-meta" &&
		min.Action != "nop" {
		return Envelope{}, fmt.Errorf(
			"invalid action: allowed 'pipeline', 'input-pipeline', " +
				"'validate', 'create-meta', 'update-meta', 'diff-meta', or 'nop'",
		)
	}
	return out, nil
}

type ErrMissingConfigPath struct{}

func (ErrMissingConfigPath) Error() string { return "missing required meta.configPath" }

func init() { Register("validate-config", ValidateConfig) }
