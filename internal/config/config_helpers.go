package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// compileCUE loads and compiles a CUE file at the given path.
// It preserves the original error messages used by callers.
func compileCUE(path string) (cue.Value, error) {
	if filepath.Ext(path) != ".cue" {
		return cue.Value{}, errors.New("unsupported config format: expected .cue")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to read config: %w", err)
	}
	ctx := cuecontext.New()
	v := ctx.CompileBytes(data)
	if err := v.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("invalid config: %v", err)
	}
	return v, nil
}
