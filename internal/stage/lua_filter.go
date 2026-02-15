package stage

import (
	"context"
	"sync"
)

func luaFilterRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine predicate
	pred := buildLuaPredicate(in)

	out := in
	out.Records = out.Records[:0]

	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	keeps := make([]bool, n)
	outs := make([]any, n)
	var envErrs []Error
	var mu sync.Mutex
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan luaFilterRes)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			keep, outAny, envE, fatal := processLuaFilterRecord(r, pred, mode)
			results <- luaFilterRes{idx: idx, keep: keep, out: outAny, envE: envE, fatal: fatal}
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
		if rr.keep {
			keeps[rr.idx] = true
			outs[rr.idx] = rr.out
		}
	}
	wg.Wait()
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	for i := 0; i < n; i++ {
		if keeps[i] {
			out.Records = append(out.Records, outs[i])
		}
	}
	return out, nil
}

func init() { Register("lua-filter", luaFilterRunner) }
