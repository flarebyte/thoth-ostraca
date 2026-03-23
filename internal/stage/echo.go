// File Guide for dev/ai agents:
// Purpose: Provide the smallest no-op stage used to verify stage wiring and basic envelope mutation.
// Responsibilities:
// - Set meta.stage to a stable marker value.
// - Return the input envelope unchanged apart from that marker.
// - Register the echo stage for contracts and smoke tests.
// Architecture notes:
// - This stage is intentionally trivial; it exists as a safe baseline for pipeline wiring and registry behavior.
// - The pure Echo helper is kept separate from the runner so tests can exercise the behavior without stage infrastructure.
package stage

import "context"

// Echo returns the same envelope, adding meta.stage = "echo".
func Echo(in Envelope) Envelope {
	if in.Meta == nil {
		in.Meta = &Meta{}
	}
	in.Meta.Stage = "echo"
	return in
}

func echoRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	return Echo(in), nil
}

func init() {
	Register("echo", echoRunner)
}
