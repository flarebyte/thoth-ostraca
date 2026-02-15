package stage

import (
	"context"
	"sync"
)

func shellExecRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// If not enabled, passthrough
	opts := buildShellOptions(in)
	if !opts.enabled {
		return in, nil
	}

	out := in
	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	type res = shellExecRes
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan res)
	var wg sync.WaitGroup
	var envErrs []Error
	var mu sync.Mutex
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			rec, envE, fatal := processShellRecord(ctx, r, opts, mode)
			results <- res{idx: idx, rec: rec, envE: envE, fatal: fatal}
		}
	}
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
	for i := 0; i < n; i++ {
		rr := <-results
		if rr.envE != nil {
			mu.Lock()
			envErrs = append(envErrs, *rr.envE)
			mu.Unlock()
		}
		if rr.fatal != nil && firstErr == nil {
			firstErr = rr.fatal
		}
		out.Records[rr.idx] = rr.rec
	}
	wg.Wait()
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	return out, nil
}

func init() { Register("shell-exec", shellExecRunner) }
