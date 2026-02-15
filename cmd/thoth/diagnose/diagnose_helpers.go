package diagnose

import (
	"context"
	"fmt"
	"os"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
	"github.com/spf13/cobra"
)

// runDiagnoseWithIn handles the mode where --in is provided.
func runDiagnoseWithIn() error {
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

// runDiagnoseWithPrepare handles the mode where --prepare is set (and --in omitted).
func runDiagnoseWithPrepare() error {
	var firstStage string
	switch flagPrepare {
	case "meta-files":
		firstStage = "discover-meta-files"
	case "input-files":
		firstStage = "discover-input-files"
	default:
		return fmt.Errorf("invalid prepare mode: %s", flagPrepare)
	}

	inEnv := stage.Envelope{Records: []stage.Record{}}
	usedRoot := relativizeRoot(flagRoot)
	inEnv.Meta = &stage.Meta{Discovery: &stage.DiscoveryMeta{Root: usedRoot, NoGitignore: flagNoGit}}
	if flagOut != "-" || flagPretty || flagLines {
		inEnv.Meta.Output = &stage.OutputMeta{Out: flagOut, Pretty: flagPretty, Lines: flagLines}
	}

	prepOut, err := stage.Run(context.Background(), firstStage, inEnv, stage.Deps{})
	if err != nil {
		return err
	}
	if flagDumpIn != "" {
		if err := writeJSONFile(flagDumpIn, prepOut); err != nil {
			return err
		}
	}

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

// runDiagnoseDefault handles the default mode: neither --in nor --prepare.
func runDiagnoseDefault(cmd *cobra.Command) error {
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
}
