package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// LoadAndValidate loads a CUE file and validates the minimal required schema.
// Required fields:
//   - configVersion: string
//   - action: string
func LoadAndValidate(path string) error {
	if filepath.Ext(path) != ".cue" {
		return errors.New("unsupported config format: expected .cue")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	ctx := cuecontext.New()
	v := ctx.CompileBytes(data)
	if err := v.Err(); err != nil {
		return fmt.Errorf("invalid config: %v", err)
	}

	// Validate required fields and types.
	if err := requireStringField(v, "configVersion"); err != nil {
		return err
	}
	if err := requireStringField(v, "action"); err != nil {
		return err
	}
	return nil
}

func requireStringField(v cue.Value, name string) error {
	f := v.LookupPath(cue.ParsePath(name))
	if !f.Exists() {
		return fmt.Errorf("missing required field: %s", name)
	}
	if f.Kind() != cue.StringKind {
		return fmt.Errorf("invalid type for field: %s (expected string)", name)
	}
	return nil
}
