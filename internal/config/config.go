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

// Minimal holds the small subset of config we validate in Phase 1.
type Minimal struct {
	ConfigVersion string
	Action        string
	Discovery     Discovery
	Filter        Filter
}

// ParseMinimal validates and extracts minimal values from the CUE config.
func ParseMinimal(path string) (Minimal, error) {
	if filepath.Ext(path) != ".cue" {
		return Minimal{}, errors.New("unsupported config format: expected .cue")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Minimal{}, fmt.Errorf("failed to read config: %w", err)
	}
	ctx := cuecontext.New()
	v := ctx.CompileBytes(data)
	if err := v.Err(); err != nil {
		return Minimal{}, fmt.Errorf("invalid config: %v", err)
	}
	if err := requireStringField(v, "configVersion"); err != nil {
		return Minimal{}, err
	}
	if err := requireStringField(v, "action"); err != nil {
		return Minimal{}, err
	}
	mv := v.LookupPath(cue.ParsePath("configVersion"))
	av := v.LookupPath(cue.ParsePath("action"))
	var m Minimal
	if err := mv.Decode(&m.ConfigVersion); err != nil {
		return Minimal{}, fmt.Errorf("invalid value for configVersion: %v", err)
	}
	if err := av.Decode(&m.Action); err != nil {
		return Minimal{}, fmt.Errorf("invalid value for action: %v", err)
	}
	// Optional discovery fields
	dv := v.LookupPath(cue.ParsePath("discovery"))
	if dv.Exists() {
		rv := dv.LookupPath(cue.ParsePath("root"))
		if rv.Exists() && rv.Kind() == cue.StringKind {
			if err := rv.Decode(&m.Discovery.Root); err == nil {
				m.Discovery.HasRoot = true
			}
		}
		ngv := dv.LookupPath(cue.ParsePath("noGitignore"))
		if ngv.Exists() && (ngv.Kind() == cue.BoolKind) {
			if err := ngv.Decode(&m.Discovery.NoGitignore); err == nil {
				m.Discovery.HasNoGitignore = true
			}
		}
	}
	// Optional filter.inline
	fv := v.LookupPath(cue.ParsePath("filter"))
	if fv.Exists() {
		iv := fv.LookupPath(cue.ParsePath("inline"))
		if iv.Exists() && iv.Kind() == cue.StringKind {
			if err := iv.Decode(&m.Filter.Inline); err == nil {
				m.Filter.HasInline = true
			}
		}
	}
	return m, nil
}

// Discovery holds optional discovery config and presence flags.
type Discovery struct {
	Root           string
	NoGitignore    bool
	HasRoot        bool
	HasNoGitignore bool
}

// Filter holds optional filter config.
type Filter struct {
	Inline    string
	HasInline bool
}
