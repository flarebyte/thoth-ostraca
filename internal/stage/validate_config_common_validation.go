package stage

import (
	"fmt"

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
	return nil
}
