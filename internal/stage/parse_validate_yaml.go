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
	if mode != "keep-going" {
		for _, rec := range in.Records {
			kv, _, fatal := processYAMLRecord(rec, root, mode)
			if fatal != nil {
				return Envelope{}, fatal
			}
			if kv.locator != "" || kv.meta != nil {
				outs = append(outs, kv)
			}
		}
		sort.Slice(outs, func(i, j int) bool { return outs[i].locator < outs[j].locator })
		out := in
		out.Records = make([]Record, 0, len(outs))
		for _, pr := range outs {
			out.Records = append(out.Records, Record{Locator: pr.locator, Meta: pr.meta})
		}
		return out, nil
	}

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
	failedByPath := map[string]string{}
	for i := 0; i < len(in.Records); i++ {
		rr := <-results
		if rr.envE != nil {
			envErrs = append(envErrs, *rr.envE)
			failedByPath[rr.envE.Locator] = rr.envE.Message
		}
		if rr.kv.meta != nil {
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
	out.Records = make([]Record, 0, len(outs)+len(failedByPath))
	for _, pr := range outs {
		out.Records = append(out.Records, Record{Locator: pr.locator, Meta: pr.meta})
	}
	for locator, msg := range failedByPath {
		out.Records = append(out.Records, Record{
			Locator: locator,
			Error:   &RecError{Stage: parseValidateYAMLStage, Message: msg},
		})
	}
	sort.Slice(out.Records, func(i, j int) bool { return out.Records[i].Locator < out.Records[j].Locator })
	return out, nil
}

func init() { Register("parse-validate-yaml", parseValidateYAMLRunner) }
