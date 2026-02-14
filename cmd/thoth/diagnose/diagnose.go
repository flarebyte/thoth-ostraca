package diagnose

import (
	"encoding/json"
	"errors"
	"fmt"
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

		var inEnv stage.Envelope
		if flagIn != "" {
			b, err := os.ReadFile(flagIn)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			if err := json.Unmarshal(b, &inEnv); err != nil {
				return fmt.Errorf("invalid input JSON: %v", err)
			}
		} else {
			inEnv = stage.Envelope{Records: []any{}}
		}

		// dump-in if requested
		if flagDumpIn != "" {
			if err := writeJSONFile(flagDumpIn, inEnv); err != nil {
				return err
			}
		}

		var outEnv stage.Envelope
		switch flagStage {
		case "echo":
			outEnv = stage.Echo(inEnv)
		default:
			return fmt.Errorf("unknown stage: %s", flagStage)
		}

		// dump-out if requested
		if flagDumpOut != "" {
			if err := writeJSONFile(flagDumpOut, outEnv); err != nil {
				return err
			}
		}

		// Print single-line JSON to stdout
		b, err := json.Marshal(outEnv)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(os.Stdout, string(b)); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&flagStage, "stage", "", "Stage name (required)")
	Cmd.Flags().StringVar(&flagIn, "in", "", "Path to input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpIn, "dump-in", "", "Path to write resolved input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpOut, "dump-out", "", "Path to write output envelope JSON")
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
