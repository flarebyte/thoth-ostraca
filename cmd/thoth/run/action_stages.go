// File Guide for dev/ai agents:
// Purpose: Define the stage order for each `thoth run` action so the CLI can build deterministic execution pipelines from validated config.
// Responsibilities:
// - Map each supported action name to its ordered list of stage names.
// - Enable or skip optional stages based on runtime metadata flags.
// - Keep action wiring decisions centralized and explicit.
// Architecture notes:
// - Action stage order is intentionally hard-coded here so changes to user-visible workflow require a deliberate review rather than emerging from config shape alone.
// - Helper predicates keep this file focused on orchestration decisions rather than repeating metadata field checks inline.
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
	case "input-pipeline":
		stages := []string{"discover-input-files"}
		if fileInfoEnabled(meta) {
			stages = append(stages, "enrich-fileinfo")
		}
		if gitEnabled(meta) {
			stages = append(stages, "enrich-git")
		}
		stages = append(stages,
			"lua-filter",
			"lua-map",
			"shell-exec",
			"lua-postmap",
		)
		if reduceEnabled(meta) {
			stages = append(stages, "lua-reduce")
		}
		if persistMetaEnabled(meta) {
			stages = append(stages,
				"load-existing-meta",
				"merge-meta",
				"write-updated-meta-files",
			)
		}
		stages = append(stages, "write-output")
		return stages, nil
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
		if filterEnabled(meta) {
			stages = append(stages, "lua-filter")
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
		if filterEnabled(meta) {
			stages = append(stages, "lua-filter")
		}
		stages = append(stages, "load-existing-meta", "merge-meta", "write-updated-meta-files", "write-output")
		return stages, nil
	case "diff-meta":
		stages := []string{
			"discover-input-files",
		}
		if filterEnabled(meta) {
			stages = append(stages, "lua-filter")
		}
		stages = append(stages,
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"compute-meta-diff",
			"write-output",
		)
		return stages, nil
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

func persistMetaEnabled(meta *stage.Meta) bool {
	return meta != nil && meta.PersistMeta != nil && meta.PersistMeta.Enabled
}

func reduceEnabled(meta *stage.Meta) bool {
	return meta != nil &&
		meta.Lua != nil &&
		meta.Lua.ReduceInline != ""
}

func filterEnabled(meta *stage.Meta) bool {
	return meta != nil &&
		meta.Lua != nil &&
		meta.Lua.FilterInline != ""
}
