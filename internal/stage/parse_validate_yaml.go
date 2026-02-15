package stage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

func parseValidateYAMLRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine root
	root := "."
	if in.Meta != nil && in.Meta.Discovery != nil && in.Meta.Discovery.Root != "" {
		root = in.Meta.Discovery.Root
	}

	// Collect output
	type kv struct {
		locator string
		meta    map[string]any
	}
	outs := make([]kv, 0, len(in.Records))
	mode, _ := errorMode(in.Meta)
	type res struct {
		kv    kv
		envE  *Error
		fatal error
	}
	workers := getWorkers(in.Meta)
	jobs := make(chan any)
	results := make(chan res)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for item := range jobs {
			r := item
			var locator string
			switch rec := r.(type) {
			case Record:
				locator = rec.Locator
			case map[string]any:
				locVal, ok := rec["locator"]
				if !ok {
					results <- res{fatal: errors.New("invalid input record: missing locator")}
					continue
				}
				s, ok := locVal.(string)
				if !ok || s == "" {
					results <- res{fatal: errors.New("invalid input record: locator must be string")}
					continue
				}
				locator = s
			default:
				results <- res{fatal: errors.New("invalid input record: expected object")}
				continue
			}
			p := filepath.Join(root, filepath.FromSlash(locator))
			b, err := os.ReadFile(p)
			if err != nil {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: fmt.Sprintf("read error: %v", err)}}
					continue
				}
				results <- res{fatal: fmt.Errorf("read error %s: %w", p, err)}
				continue
			}
			var y any
			if err := yaml.Unmarshal(b, &y); err != nil {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: fmt.Sprintf("invalid YAML: %v", err)}}
					continue
				}
				results <- res{fatal: fmt.Errorf("invalid YAML %s: %v", p, err)}
				continue
			}
			ym, ok := y.(map[string]any)
			if !ok {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: "top-level must be mapping"}}
					continue
				}
				results <- res{fatal: fmt.Errorf("invalid YAML %s: top-level must be mapping", p)}
				continue
			}
			yloc, ok := ym["locator"]
			if !ok {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: "missing required field: locator"}}
					continue
				}
				results <- res{fatal: fmt.Errorf("invalid YAML %s: missing required field: locator", p)}
				continue
			}
			ylocStr, ok := yloc.(string)
			if !ok || ylocStr == "" {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: "invalid type for field: locator"}}
					continue
				}
				results <- res{fatal: fmt.Errorf("invalid YAML %s: invalid type for field: locator", p)}
				continue
			}
			ymeta, ok := ym["meta"]
			if !ok {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: "missing required field: meta"}}
					continue
				}
				results <- res{fatal: fmt.Errorf("invalid YAML %s: missing required field: meta", p)}
				continue
			}
			ymetaMap, ok := ymeta.(map[string]any)
			if !ok {
				if mode == "keep-going" {
					results <- res{kv: kv{locator: locator, meta: nil}, envE: &Error{Stage: "parse-validate-yaml", Locator: locator, Message: "invalid type for field: meta"}}
					continue
				}
				results <- res{fatal: fmt.Errorf("invalid YAML %s: invalid type for field: meta", p)}
				continue
			}
			results <- res{kv: kv{locator: ylocStr, meta: ymetaMap}}
		}
	}
	// start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		for _, r := range in.Records {
			jobs <- r
		}
		close(jobs)
	}()
	var firstErr error
	var envErrs []Error
	for i := 0; i < len(in.Records); i++ {
		rr := <-results
		if rr.envE != nil {
			envErrs = append(envErrs, *rr.envE)
		}
		if rr.kv.locator != "" || rr.kv.meta != nil {
			outs = append(outs, rr.kv)
		}
		if firstErr == nil && rr.fatal != nil {
			firstErr = rr.fatal
		}
	}
	wg.Wait()
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	// Append collected errors
	if len(envErrs) > 0 {
		outE := in
		outE.Errors = append(outE.Errors, envErrs...)
		in = outE
	}

	sort.Slice(outs, func(i, j int) bool { return outs[i].locator < outs[j].locator })

	out := in
	out.Records = make([]any, 0, len(outs))
	for _, pr := range outs {
		rec := Record{Locator: pr.locator}
		if pr.meta != nil {
			rec.Meta = pr.meta
		} else {
			// In keep-going with error, embed if requested
			rec.Error = &RecError{Stage: "parse-validate-yaml", Message: "failed"}
		}
		out.Records = append(out.Records, rec)
	}
	return out, nil
}

func init() { Register("parse-validate-yaml", parseValidateYAMLRunner) }
