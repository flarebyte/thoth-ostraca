package diagnose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	runpkg "github.com/flarebyte/thoth-ostraca/cmd/thoth/run"
	"github.com/flarebyte/thoth-ostraca/internal/stage"
	"github.com/spf13/cobra"
)

func maybeDumpIn(env stage.Envelope) error {
	if flagDumpDir != "" || flagDumpIn == "" {
		return nil
	}
	return writeJSONFile(flagDumpIn, env)
}

func maybeDumpOut(env stage.Envelope) error {
	if flagDumpDir != "" || flagDumpOut == "" {
		return nil
	}
	return writeJSONFile(flagDumpOut, env)
}

func dumpStageBoundary(seq int, stageName string, suffix string, env stage.Envelope) error {
	if flagDumpDir == "" {
		return nil
	}
	base := fmt.Sprintf("%03d_%s_%s.json", seq, stageName, suffix)
	return writeJSONFile(filepath.Join(flagDumpDir, base), env)
}

func runStagesAndRender(inEnv stage.Envelope, stages []string) error {
	outEnv, err := runStageSequence(inEnv, stages)
	if err != nil {
		return err
	}
	if outEnv.Meta == nil {
		outEnv.Meta = &stage.Meta{}
	}
	outEnv.Meta.ContractVersion = "1"
	if err := maybeDumpOut(outEnv); err != nil {
		return err
	}
	return printEnvelopeOneLine(os.Stdout, outEnv)
}

func runStageSequence(inEnv stage.Envelope, stages []string) (stage.Envelope, error) {
	out := inEnv
	for i, stageName := range stages {
		seq := i + 1
		if err := dumpStageBoundary(seq, stageName, "in", out); err != nil {
			return stage.Envelope{}, err
		}
		next, err := stage.Run(context.Background(), stageName, out, stage.Deps{})
		if err != nil {
			return stage.Envelope{}, err
		}
		if err := dumpStageBoundary(seq, stageName, "out", next); err != nil {
			return stage.Envelope{}, err
		}
		out = next
	}
	return out, nil
}

func resolveStageByIndex(stages []string, idx int, flagName string) (string, error) {
	if idx < 0 || idx >= len(stages) {
		return "", fmt.Errorf("%s out of range: %d", flagName, idx)
	}
	return stages[idx], nil
}

func findStageIndexByName(stages []string, stageName string, flagName string) (int, error) {
	for i, cur := range stages {
		if cur == stageName {
			return i, nil
		}
	}
	return -1, fmt.Errorf("unknown %s: %s", flagName, stageName)
}

func resolveTargetStage(stages []string) (string, error) {
	if flagStageIndex >= 0 {
		return resolveStageByIndex(stages, flagStageIndex, "--stage-index")
	}
	if flagStage == "" {
		return "", fmt.Errorf("missing required flag: --stage")
	}
	if _, err := findStageIndexByName(stages, flagStage, "--stage"); err != nil {
		return "", err
	}
	return flagStage, nil
}

func resolveUntilIndex(stages []string) (int, bool, error) {
	if flagUntilIndex >= 0 {
		if flagUntilIndex < 0 || flagUntilIndex >= len(stages) {
			return -1, false, fmt.Errorf("--until-index out of range: %d", flagUntilIndex)
		}
		return flagUntilIndex, true, nil
	}
	if flagUntilStage == "" {
		return -1, false, nil
	}
	idx, err := findStageIndexByName(stages, flagUntilStage, "--until-stage")
	if err != nil {
		return -1, false, err
	}
	return idx, true, nil
}

func resolvePreparedAction() (string, error) {
	switch flagPreparePipeline {
	case "pipeline", "validate", "create-meta", "update-meta", "diff-meta":
		return flagPreparePipeline, nil
	default:
		return "", fmt.Errorf("invalid prepare-pipeline action: %s", flagPreparePipeline)
	}
}

func runDiagnoseWithIn() error {
	inEnv, err := prepareDiagnoseInput(flagIn, flagConfig, "", false)
	if err != nil {
		return err
	}
	if err := maybeDumpIn(inEnv); err != nil {
		return err
	}
	stageName := flagStage
	if stageName == "" {
		return fmt.Errorf("missing required flag: --stage")
	}
	return runStagesAndRender(inEnv, []string{stageName})
}

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

	prepOut, err := runStageSequence(inEnv, []string{firstStage})
	if err != nil {
		return err
	}
	if err := maybeDumpIn(prepOut); err != nil {
		return err
	}
	return runStagesAndRender(prepOut, []string{flagStage})
}

func runDiagnoseWithPreparedPipeline(cmd *cobra.Command) error {
	action, err := resolvePreparedAction()
	if err != nil {
		return err
	}
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
	if flagConfig != "" {
		validated, err := stage.Run(context.Background(), "validate-config", inEnv, stage.Deps{})
		if err != nil {
			return err
		}
		inEnv = validated
	}

	stages, err := runpkg.PreparedActionStages(action, inEnv.Meta)
	if err != nil {
		return err
	}
	if len(stages) == 0 {
		return fmt.Errorf("empty prepared pipeline")
	}

	if err := maybeDumpIn(inEnv); err != nil {
		return err
	}

	untilIdx, hasUntil, err := resolveUntilIndex(stages)
	if err != nil {
		return err
	}
	if hasUntil {
		return runStagesAndRender(inEnv, stages[:untilIdx+1])
	}

	target, err := resolveTargetStage(stages)
	if err != nil {
		return err
	}
	return runStagesAndRender(inEnv, []string{target})
}

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
	if err := maybeDumpIn(inEnv); err != nil {
		return err
	}
	stageName := strings.TrimSpace(flagStage)
	if stageName == "" {
		return fmt.Errorf("missing required flag: --stage")
	}
	return runStagesAndRender(inEnv, []string{stageName})
}
