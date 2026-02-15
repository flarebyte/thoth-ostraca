package diagnose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
	"github.com/spf13/cobra"
)

var (
	flagStage   string
	flagIn      string
	flagDumpIn  string
	flagDumpOut string
	flagPrepare string
	flagConfig  string
	flagRoot    string
	flagNoGit   bool
	flagOut     string
	flagPretty  bool
	flagLines   bool
)

// Cmd implements `thoth diagnose`.
var Cmd = &cobra.Command{
	Use:           "diagnose",
	Short:         "Run a single diagnostic stage",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagStage == "" {
			return errors.New("missing required flag: --stage")
		}
		// If --in is provided, it takes precedence and --prepare is ignored.
		if flagIn != "" {
			inEnv, err := prepareDiagnoseInput(flagIn, flagConfig, "", false)
			if err != nil {
				return err
			}
			if flagDumpIn != "" {
				if err := writeJSONFile(flagDumpIn, inEnv); err != nil {
					return err
				}
			}
			outEnv, err := stage.Run(context.Background(), flagStage, inEnv, stage.Deps{})
			if err != nil {
				return err
			}
			if outEnv.Meta == nil {
				outEnv.Meta = &stage.Meta{}
			}
			outEnv.Meta.ContractVersion = "1"
			if flagDumpOut != "" {
				if err := writeJSONFile(flagDumpOut, outEnv); err != nil {
					return err
				}
			}
			return printEnvelopeOneLine(os.Stdout, outEnv)
		}

		// If --prepare is set (and --in omitted), run the corresponding discovery stage first.
		if flagPrepare != "" {
			var firstStage string
			switch flagPrepare {
			case "meta-files":
				firstStage = "discover-meta-files"
			case "input-files":
				firstStage = "discover-input-files"
			default:
				return fmt.Errorf("invalid prepare mode: %s", flagPrepare)
			}
			// Build minimal env with discovery from flags (root defaults to ".").
			inEnv := stage.Envelope{Records: []stage.Record{}}
			usedRoot := relativizeRoot(flagRoot)
			inEnv.Meta = &stage.Meta{Discovery: &stage.DiscoveryMeta{Root: usedRoot, NoGitignore: flagNoGit}}
			// Optional output when diagnosing without --in; only if flags deviate from defaults
			if flagOut != "-" || flagPretty || flagLines {
				inEnv.Meta.Output = &stage.OutputMeta{Out: flagOut, Pretty: flagPretty, Lines: flagLines}
			}
			// Run preparation stage
			prepOut, err := stage.Run(context.Background(), firstStage, inEnv, stage.Deps{})
			if err != nil {
				return err
			}
			if flagDumpIn != "" {
				if err := writeJSONFile(flagDumpIn, prepOut); err != nil {
					return err
				}
			}
			// Execute target stage using prepared input
			outEnv, err := stage.Run(context.Background(), flagStage, prepOut, stage.Deps{})
			if err != nil {
				return err
			}
			if outEnv.Meta == nil {
				outEnv.Meta = &stage.Meta{}
			}
			outEnv.Meta.ContractVersion = "1"
			if flagDumpOut != "" {
				if err := writeJSONFile(flagDumpOut, outEnv); err != nil {
					return err
				}
			}
			return printEnvelopeOneLine(os.Stdout, outEnv)
		}

		// Neither --in nor --prepare: keep existing behavior.
		// Only apply discovery flags when explicitly provided by the user.
		changedRoot := cmd.Flags().Changed("root")
		changedNoGit := cmd.Flags().Changed("no-gitignore")
		baseRoot := ""
		baseNoGit := false
		if changedRoot {
			baseRoot = relativizeRoot(flagRoot)
		}
		if changedNoGit {
			baseNoGit = flagNoGit
		}
		inEnv, err := prepareDiagnoseInput("", flagConfig, baseRoot, baseNoGit)
		if err != nil {
			return err
		}
		if flagDumpIn != "" {
			if err := writeJSONFile(flagDumpIn, inEnv); err != nil {
				return err
			}
		}
		outEnv, err := stage.Run(context.Background(), flagStage, inEnv, stage.Deps{})
		if err != nil {
			return err
		}
		if outEnv.Meta == nil {
			outEnv.Meta = &stage.Meta{}
		}
		outEnv.Meta.ContractVersion = "1"
		if flagDumpOut != "" {
			if err := writeJSONFile(flagDumpOut, outEnv); err != nil {
				return err
			}
		}
		return printEnvelopeOneLine(os.Stdout, outEnv)
	},
}

// relativizeRoot converts an absolute root under the current working directory
// to a relative path for deterministic output; otherwise returns the input.
func relativizeRoot(root string) string {
	if root == "" || root == "." {
		return root
	}
	if !filepath.IsAbs(root) {
		// Normalize separators for JSON stability
		return filepath.ToSlash(root)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return root
	}
	rel, err := filepath.Rel(cwd, root)
	if err != nil {
		return root
	}
	if len(rel) == 0 || rel == "." || rel == root || rel[:2] == ".." {
		return root
	}
	return filepath.ToSlash(rel)
}

func init() {
	Cmd.Flags().StringVar(&flagStage, "stage", "", "Stage name (required)")
	Cmd.Flags().StringVar(&flagIn, "in", "", "Path to input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpIn, "dump-in", "", "Path to write resolved input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpOut, "dump-out", "", "Path to write output envelope JSON")
	Cmd.Flags().StringVar(&flagPrepare, "prepare", "", "Prepare input via discovery: meta-files|input-files")
	Cmd.Flags().StringVar(&flagConfig, "config", "", "Config path used when --in omitted")
	Cmd.Flags().StringVar(&flagRoot, "root", ".", "Discovery root (prepare mode)")
	Cmd.Flags().BoolVar(&flagNoGit, "no-gitignore", false, "Disable .gitignore (prepare mode)")
	Cmd.Flags().StringVar(&flagOut, "out", "-", "Output path for write-output (diagnose --in omitted)")
	Cmd.Flags().BoolVar(&flagPretty, "pretty", false, "Pretty JSON (diagnose --in omitted)")
	Cmd.Flags().BoolVar(&flagLines, "lines", false, "Lines mode (diagnose --in omitted)")
}

func writeJSONFile(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create dump dir: %w", err)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func prepareDiagnoseInput(inPath, cfg, root string, noGit bool) (stage.Envelope, error) {
	if inPath != "" {
		b, err := os.ReadFile(inPath)
		if err != nil {
			return stage.Envelope{}, fmt.Errorf("failed to read input: %w", err)
		}
		var env stage.Envelope
		if err := json.Unmarshal(b, &env); err != nil {
			return stage.Envelope{}, fmt.Errorf("invalid input JSON: %v", err)
		}
		return env, nil
	}
	env := stage.Envelope{Records: []stage.Record{}}
	if cfg != "" {
		env.Meta = &stage.Meta{ConfigPath: cfg}
	}
	if root != "" || noGit {
		if env.Meta == nil {
			env.Meta = &stage.Meta{}
		}
		env.Meta.Discovery = &stage.DiscoveryMeta{}
		if root != "" {
			env.Meta.Discovery.Root = root
		}
		if noGit {
			env.Meta.Discovery.NoGitignore = true
		}
	}
	// Optional output when diagnosing without --in; only if flags deviate from defaults
	if flagOut != "-" || flagPretty || flagLines {
		if env.Meta == nil {
			env.Meta = &stage.Meta{}
		}
		env.Meta.Output = &stage.OutputMeta{Out: flagOut, Pretty: flagPretty, Lines: flagLines}
	}
	return env, nil
}

func printEnvelopeOneLine(w io.Writer, env stage.Envelope) error {
	if env.Meta == nil {
		env.Meta = &stage.Meta{}
	}
	env.Meta.ContractVersion = "1"
	stage.SortEnvelopeErrors(&env)
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, string(b)); err != nil {
		return err
	}
	return nil
}
