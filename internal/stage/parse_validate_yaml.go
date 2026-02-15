package stage

import (
	"context"
	"sort"
	"sync"
)

func parseValidateYAMLRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine root
	root := determineRoot(in)

	// Collect output
	outs := make([]yamlKV, 0, len(in.Records))
	mode, _ := errorMode(in.Meta)
	type res struct {
		kv    yamlKV
		envE  *Error
		fatal error
	}
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan res)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for item := range jobs {
			rec := in.Records[item]
			kv, envE, fatal := processYAMLRecord(rec, root, mode)
			results <- res{kv: kv, envE: envE, fatal: fatal}
		}
	}
	// start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		for i := range in.Records {
			jobs <- i
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
	out.Records = make([]Record, 0, len(outs))
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
