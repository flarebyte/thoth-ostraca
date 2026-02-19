package stage

import (
	"context"
	"errors"
)

// validate top-level keys of the generic JSON map.
func validateTopLevel(g map[string]any) error {
	allowedTop := map[string]bool{"records": true, "meta": true, "errors": true}
	for k := range g {
		if !allowedTop[k] {
			return errors.New("unexpected top-level key: " + k)
		}
	}
	return nil
}

// validateRecordsSection asserts the shape of the records array.
func validateRecordsSection(v any) error {
	recs, ok := v.([]any)
	if !ok {
		return errors.New("records must be array")
	}
	for _, it := range recs {
		m, ok := it.(map[string]any)
		if !ok {
			return errors.New("record must be object")
		}
		if err := validateRecord(m); err != nil {
			return err
		}
	}
	return nil
}

// validateRecord validates a single record map.
func validateRecord(m map[string]any) error {
	allowedRec := map[string]bool{
		"locator": true, "meta": true, "mapped": true, "shell": true, "post": true, "error": true, "fileInfo": true, "git": true,
	}
	for k := range m {
		if !allowedRec[k] {
			return errors.New("unexpected record key: " + k)
		}
	}
	if loc, ok := m["locator"]; ok {
		if _, ok := loc.(string); !ok {
			return errors.New("record.locator must be string")
		}
	}
	if errv, hasErr := m["error"]; hasErr && errv != nil {
		em, ok := errv.(map[string]any)
		if !ok {
			return errors.New("record.error must be object")
		}
		if _, ok := em["stage"].(string); !ok {
			return errors.New("record.error.stage must be string")
		}
		if _, ok := em["message"].(string); !ok {
			return errors.New("record.error.message must be string")
		}
		if loc, ok := em["locator"]; ok && loc != nil {
			if _, ok := loc.(string); !ok {
				return errors.New("record.error.locator must be string")
			}
		}
		for k := range em {
			if k != "stage" && k != "message" && k != "locator" {
				return errors.New("unexpected error key: " + k)
			}
		}
	}
	return nil
}

// selectStages returns the ordered stage list for a given action.
func selectStages(action string) ([]string, error) {
	switch action {
	case "pipeline", "nop":
		return []string{"discover-meta-files", "parse-validate-yaml", "validate-locators", "lua-filter", "lua-map", "shell-exec", "lua-postmap", "lua-reduce"}, nil
	case "validate":
		return []string{"discover-meta-files", "parse-validate-yaml", "validate-locators"}, nil
	case "create-meta":
		return []string{"discover-input-files", "enrich-fileinfo", "enrich-git", "write-meta-files"}, nil
	case "update-meta":
		return []string{"discover-input-files", "enrich-fileinfo", "enrich-git", "load-existing-meta", "merge-meta", "write-updated-meta-files"}, nil
	case "diff-meta":
		return []string{"discover-input-files", "discover-meta-files", "parse-validate-yaml", "validate-locators", "compute-meta-diff"}, nil
	default:
		return nil, errors.New("unknown action: " + action)
	}
}

// runStagesTest executes the stages in order starting from the given envelope.
func runStagesTest(ctx context.Context, stages []string, start Envelope) (Envelope, string, error) {
	cur := start
	var err error
	for _, s := range stages {
		cur, err = Run(ctx, s, cur, Deps{})
		if err != nil {
			return Envelope{}, s, err
		}
	}
	return cur, "", nil
}
