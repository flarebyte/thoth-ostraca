package diagnose

import (
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
	flagStage           string
	flagStageIndex      int
	flagUntilStage      string
	flagUntilIndex      int
	flagIn              string
	flagDumpIn          string
	flagDumpOut         string
	flagDumpDir         string
	flagPrepare         string
	flagPreparePipeline string
	flagConfig          string
	flagRoot            string
	flagNoGit           bool
	flagOut             string
	flagPretty          bool
	flagLines           bool
)

// Cmd implements `thoth diagnose`.
var Cmd = &cobra.Command{
	Use:           "diagnose",
	Short:         "Run a single diagnostic stage",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagIn != "" {
			return runDiagnoseWithIn()
		}
		if flagPreparePipeline != "" {
			return runDiagnoseWithPreparedPipeline(cmd)
		}
		if flagPrepare != "" {
			if flagStage == "" {
				return errors.New("missing required flag: --stage")
			}
			return runDiagnoseWithPrepare()
		}
		if flagStage == "" {
			return errors.New("missing required flag: --stage")
		}
		return runDiagnoseDefault(cmd)
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
	Cmd.Flags().IntVar(&flagStageIndex, "stage-index", -1, "Stage index in prepared pipeline (0-based)")
	Cmd.Flags().StringVar(&flagUntilStage, "until-stage", "", "Run prepared pipeline through this stage name (inclusive)")
	Cmd.Flags().IntVar(&flagUntilIndex, "until-index", -1, "Run prepared pipeline through this stage index (inclusive, 0-based)")
	Cmd.Flags().StringVar(&flagIn, "in", "", "Path to input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpIn, "dump-in", "", "Path to write resolved input envelope JSON")
	Cmd.Flags().StringVar(&flagDumpOut, "dump-out", "", "Path to write output envelope JSON")
	Cmd.Flags().StringVar(&flagDumpDir, "dump-dir", "", "Directory to write per-stage dumps (<seq>_<stage>_{in,out}.json)")
	Cmd.Flags().StringVar(&flagPrepare, "prepare", "", "Prepare input via discovery: meta-files|input-files")
	Cmd.Flags().StringVar(&flagPreparePipeline, "prepare-pipeline", "", "Prepare action pipeline: pipeline|validate|create-meta|update-meta|diff-meta")
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
