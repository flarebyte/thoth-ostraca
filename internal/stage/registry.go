// File Guide for dev/ai agents:
// Purpose: Hold the stage runner registry that maps validated stage names to executable implementations.
// Responsibilities:
// - Define the shared stage runner function signature and runtime dependencies.
// - Register named stage runners during package initialization.
// - Resolve and execute stages by name with a uniform unknown-stage error.
// Architecture notes:
// - The registry is intentionally package-global and small because stage composition is built at startup, not via dynamic plugin loading.
// - Deps is kept minimal so stage implementations can share only the runtime channels and writers they actually need.
package stage

import (
	"context"
	"io"
)

// Deps is a placeholder for stage dependencies (e.g., FS, log, etc.).
// Keep minimal for now.
type Deps struct {
	RecordStream <-chan Record
	Stderr       io.Writer
}

// Runner executes a stage.
type Runner func(ctx context.Context, in Envelope, deps Deps) (Envelope, error)

var registry = map[string]Runner{}

// Register adds a stage runner.
func Register(name string, r Runner) {
	registry[name] = r
}

// Run executes a registered stage by name.
func Run(ctx context.Context, name string, in Envelope, deps Deps) (Envelope, error) {
	r, ok := registry[name]
	if !ok {
		return Envelope{}, ErrUnknown{name: name}
	}
	return r(ctx, in, deps)
}

// ErrUnknown is returned when a stage is not found.
type ErrUnknown struct{ name string }

func (e ErrUnknown) Error() string { return "unknown stage: " + e.name }
