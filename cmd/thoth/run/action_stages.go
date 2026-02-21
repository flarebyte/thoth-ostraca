package run

import (
	"fmt"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// PreparedActionStages returns the deterministic stage order used for an action.
func PreparedActionStages(action string, meta *stage.Meta) ([]string, error) {
	switch action {
	case "pipeline", "nop":
		return []string{
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"lua-filter",
			"lua-map",
			"shell-exec",
			"lua-postmap",
			"lua-reduce",
			"write-output",
		}, nil
	case "validate":
		return []string{
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"write-output",
		}, nil
	case "create-meta":
		stages := []string{"discover-input-files"}
		if fileInfoEnabled(meta) {
			stages = append(stages, "enrich-fileinfo")
		}
		if gitEnabled(meta) {
			stages = append(stages, "enrich-git")
		}
		stages = append(stages, "write-meta-files", "write-output")
		return stages, nil
	case "update-meta":
		stages := []string{"discover-input-files"}
		if fileInfoEnabled(meta) {
			stages = append(stages, "enrich-fileinfo")
		}
		if gitEnabled(meta) {
			stages = append(stages, "enrich-git")
		}
		stages = append(stages, "load-existing-meta", "merge-meta", "write-updated-meta-files", "write-output")
		return stages, nil
	case "diff-meta":
		return []string{
			"discover-input-files",
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"compute-meta-diff",
			"write-output",
		}, nil
	default:
		return nil, fmt.Errorf("invalid action")
	}
}

func fileInfoEnabled(meta *stage.Meta) bool {
	return meta != nil && meta.FileInfo != nil && meta.FileInfo.Enabled
}

func gitEnabled(meta *stage.Meta) bool {
	return meta != nil && meta.Git != nil && meta.Git.Enabled
}
