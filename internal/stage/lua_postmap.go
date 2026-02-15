package stage

import (
	"context"
	"sync"
)

func luaPostMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	mode, _ := errorMode(in.Meta)
	// If no inline, apply deterministic default without Lua to keep ordering stable
	if in.Meta == nil || in.Meta.Lua == nil || in.Meta.Lua.PostMapInline == "" {
		out := in
		n := len(in.Records)
		workers := getWorkers(in.Meta)
		jobs := make(chan int)
		results := make(chan luaPostMapRes)
		var wg sync.WaitGroup
		worker := func() {
			defer wg.Done()
			for idx := range jobs {
				r := in.Records[idx]
				rec, envE, fatal := processDefaultPostMapRecord(r, mode)
				results <- luaPostMapRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
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
				out.Errors = append(out.Errors, *rr.envE)
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
		return out, nil
	}

	code := buildLuaPostMapCode(in)

	out := in
	n := len(in.Records)
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan luaPostMapRes)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			rec, envE, fatal := processLuaPostMapRecord(r, code, mode)
			results <- luaPostMapRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
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
			out.Errors = append(out.Errors, *rr.envE)
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
	return out, nil
}

func init() { Register("lua-postmap", luaPostMapRunner) }
