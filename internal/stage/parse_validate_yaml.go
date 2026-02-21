package stage

import (
	"context"
	"sort"
	"sync"
)

func parseValidateYAMLRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	root := determineRoot(in)
	allowUnknownTop := allowUnknownTopLevel(in)
	maxBytes := maxYAMLBytes(in)
	mode, _ := errorMode(in.Meta)
	_, embed := errorMode(in.Meta)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type res struct {
		path  string
		kv    yamlKV
		envE  *Error
		fatal error
	}
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan res, len(in.Records))
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-jobs:
				if !ok {
					return
				}
				rec := in.Records[item]
				path := rec.Locator
				kv, envE, fatal := processYAMLRecord(rec, root, mode, allowUnknownTop, maxBytes)
				select {
				case results <- res{path: path, kv: kv, envE: envE, fatal: fatal}:
				case <-ctx.Done():
					return
				}
			}
		}
	}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		defer close(jobs)
		for i := range in.Records {
			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	outs := make([]yamlKV, 0, len(in.Records))
	envErrs := make([]Error, 0)
	failedRecords := make([]Record, 0)
	for rr := range results {
		if rr.fatal != nil {
			if firstErr == nil {
				firstErr = rr.fatal
				if mode != "keep-going" {
					cancel()
				}
			}
			continue
		}
		if rr.envE != nil {
			se := sanitizedError(*rr.envE)
			envErrs = append(envErrs, se)
			if mode == "keep-going" {
				fr := Record{Locator: rr.path}
				if embed {
					fr.Error = &RecError{Stage: parseValidateYAMLStage, Message: se.Message}
				}
				failedRecords = append(failedRecords, fr)
			}
			continue
		}
		if rr.kv.meta != nil {
			outs = append(outs, rr.kv)
		}
	}
	if firstErr != nil {
		return Envelope{}, firstErr
	}

	sort.Slice(outs, func(i, j int) bool { return outs[i].locator < outs[j].locator })

	out := in
	out.Records = make([]Record, 0, len(outs)+len(failedRecords))
	for _, pr := range outs {
		out.Records = append(out.Records, Record{Locator: pr.locator, Meta: pr.meta})
	}
	out.Records = append(out.Records, failedRecords...)
	sort.Slice(out.Records, func(i, j int) bool { return out.Records[i].Locator < out.Records[j].Locator })
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
		SortEnvelopeErrors(&out)
	}
	return out, nil
}

func init() { Register("parse-validate-yaml", parseValidateYAMLRunner) }
