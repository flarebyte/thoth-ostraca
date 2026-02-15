package config

import (
	"fmt"

	"cuelang.org/go/cue"
)

// LoadAndValidate loads a CUE file and validates the minimal required schema.
// Required fields:
//   - configVersion: string
//   - action: string
func LoadAndValidate(path string) error {
	v, err := compileCUE(path)
	if err != nil {
		return err
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
	Map           Map
	Shell         Shell
	PostMap       PostMap
	Reduce        Reduce
	Output        Output
	Errors        Errors
	Workers       Workers
}

// ParseMinimal validates and extracts minimal values from the CUE config.
func ParseMinimal(path string) (Minimal, error) {
	v, err := compileCUE(path)
	if err != nil {
		return Minimal{}, err
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
	// Optional sections
	m.Discovery = parseDiscoverySection(v)
	m.Filter = parseFilterSection(v)
	m.Map = parseMapSection(v)
	m.Shell = parseShellSection(v)
	m.PostMap = parsePostMapSection(v)
	m.Reduce = parseReduceSection(v)
	m.Output = parseOutputSection(v)
	m.Errors = parseErrorsSection(v)
	m.Workers = parseWorkersSection(v)
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

// Map holds optional map config.
type Map struct {
	Inline    string
	HasInline bool
}

// Shell holds optional shell execution configuration.
type Shell struct {
	Enabled      bool
	Program      string
	ArgsTemplate []string
	TimeoutMs    int
	HasEnabled   bool
	HasProgram   bool
	HasArgs      bool
	HasTimeout   bool
}

// PostMap holds optional post-map configuration.
type PostMap struct {
	Inline    string
	HasInline bool
}

// Reduce holds optional reduce config.
type Reduce struct {
	Inline    string
	HasInline bool
}

// Output holds optional output config.
type Output struct {
	Lines    bool
	HasLines bool
}

// Errors holds error handling mode config.
type Errors struct {
	Mode        string
	EmbedErrors bool
	HasMode     bool
	HasEmbed    bool
}

// Workers holds optional worker count.
type Workers struct {
	Count    int
	HasCount bool
}
