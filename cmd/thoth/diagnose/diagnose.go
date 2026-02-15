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
	flagConfig  string
	flagRoot    string
	flagNoGit   bool
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

		inEnv, err := prepareDiagnoseInput(flagIn, flagConfig, flagRoot, flagNoGit)
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
        // Attach contract version for final outputs (stdout and dump-out)
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

func init() {
	Cmd.Flags().StringVar(&flagStage, "stage", "", "Stage name (required)")
	Cmd.Flags().StringVar(&flagIn, "in", "", "Path to input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpIn, "dump-in", "", "Path to write resolved input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpOut, "dump-out", "", "Path to write output envelope JSON")
	Cmd.Flags().StringVar(&flagConfig, "config", "", "Config path used when --in omitted")
	Cmd.Flags().StringVar(&flagRoot, "root", "", "Discovery root used when --in omitted")
	Cmd.Flags().BoolVar(&flagNoGit, "no-gitignore", false, "Disable .gitignore filtering (when --in omitted)")
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
