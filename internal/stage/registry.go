package stage

import "context"

// Deps is a placeholder for stage dependencies (e.g., FS, log, etc.).
// Keep minimal for now.
type Deps struct {
	RecordStream <-chan Record
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
