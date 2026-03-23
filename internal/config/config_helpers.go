// File Guide for dev/ai agents:
// Purpose: Compile raw CUE config files into cue.Value while preserving caller-facing error shape.
// Responsibilities:
// - Read config files from disk.
// - Reject unsupported extensions before compilation.
// - Compile bytes into a cue.Value used by higher-level parsing.
// Architecture notes:
// - Error strings are intentionally simple because callers and tests depend on their wording.
// - This file does not do semantic config parsing; it only provides the compiled CUE value.
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
