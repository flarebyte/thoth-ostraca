package stage

import (
	"fmt"
	"strings"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

func validateCommonConfig(min config.Minimal) error {
	if min.Workers.HasCount && min.Workers.Count < 1 {
		return fmt.Errorf("invalid workers: must be >= 1")
	}
	if min.Limits.HasMaxRecordsInMemory && min.Limits.MaxRecordsInMemory < 1 {
		return fmt.Errorf("invalid limits.maxRecordsInMemory: must be >= 1")
	}
	if min.Shell.HasEnabled && min.Shell.Enabled && len(min.Shell.ArgsTemplate) == 0 {
		return fmt.Errorf("invalid shell.argsTemplate: required when shell.enabled=true")
	}
	if min.Shell.DecodeJSONStdout &&
		min.Shell.HasCaptureStdout &&
		!min.Shell.CaptureStdout {
		return fmt.Errorf(
			"invalid shell.capture.stdout: must be true when " +
				"shell.decodeJsonStdout=true",
		)
	}
	if min.PersistMeta.Enabled && min.Action != "input-pipeline" {
		return fmt.Errorf(
			"invalid persistMeta: only supported for action " +
				"'input-pipeline'",
		)
	}
	if min.PersistMeta.HasOutDir &&
		strings.TrimSpace(min.PersistMeta.OutDir) == "" {
		return fmt.Errorf("invalid persistMeta.outDir: must be non-empty")
	}
	if min.PersistMeta.HasOutDir && !min.PersistMeta.Enabled {
		return fmt.Errorf(
			"invalid persistMeta.outDir: requires persistMeta.enabled=true",
		)
	}
	return nil
}
