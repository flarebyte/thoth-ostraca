package stage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaPostMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	mode, _ := errorMode(in.Meta)
	// If no inline, apply deterministic default without Lua to keep ordering stable
	if in.Meta == nil || in.Meta.Lua == nil || in.Meta.Lua.PostMapInline == "" {
		out := in
		n := len(in.Records)
		type res struct {
			idx   int
			rec   any
			envE  *Error
			fatal error
		}
		workers := getWorkers(in.Meta)
		jobs := make(chan int)
		results := make(chan res)
		var wg sync.WaitGroup
		worker := func() {
			defer wg.Done()
			for idx := range jobs {
				r := in.Records[idx]
				rec, ok := r.(Record)
				if !ok {
					if mode == "keep-going" {
						results <- res{idx: idx, rec: r, envE: &Error{Stage: "lua-postmap", Message: "invalid record type"}}
						continue
					}
					results <- res{idx: idx, fatal: errors.New("lua-postmap: invalid record type")}
					continue
				}
				if rec.Error != nil {
					results <- res{idx: idx, rec: rec}
					continue
				}
				type postResult struct {
					Locator string `json:"locator"`
					Exit    int    `json:"exit,omitempty"`
				}
				pr := postResult{Locator: rec.Locator}
				if rec.Shell != nil {
					pr.Exit = rec.Shell.ExitCode
				}
				rec.Post = pr
				results <- res{idx: idx, rec: rec}
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

	code := in.Meta.Lua.PostMapInline
	if !containsReturn(code) {
		code = "return (" + code + ")"
	}

	out := in
	n := len(in.Records)
	type res struct {
		idx   int
		rec   any
		envE  *Error
		fatal error
	}
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan res)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			rec, ok := r.(Record)
			if !ok {
				if mode == "keep-going" {
					results <- res{idx: idx, rec: r, envE: &Error{Stage: "lua-postmap", Message: "invalid record type"}}
					continue
				}
				results <- res{idx: idx, fatal: errors.New("lua-postmap: invalid record type")}
				continue
			}

			L := lua.NewState(lua.Options{SkipOpenLibs: true})
			// Minimal libs
			L.Push(L.NewFunction(lua.OpenBase))
			L.Push(lua.LString("base"))
			L.Call(1, 0)
			L.Push(L.NewFunction(lua.OpenString))
			L.Push(lua.LString("string"))
			L.Call(1, 0)
			L.Push(L.NewFunction(lua.OpenTable))
			L.Push(lua.LString("table"))
			L.Call(1, 0)
			L.Push(L.NewFunction(lua.OpenMath))
			L.Push(lua.LString("math"))
			L.Call(1, 0)

			// Globals
			L.SetGlobal("locator", lua.LString(rec.Locator))
			L.SetGlobal("meta", toLValue(L, rec.Meta))
			L.SetGlobal("mapped", toLValue(L, rec.Mapped))
			var shellMap map[string]any
			if rec.Shell != nil {
				shellMap = map[string]any{
					"exitCode": rec.Shell.ExitCode,
					"stdout":   rec.Shell.Stdout,
					"stderr":   rec.Shell.Stderr,
				}
			}
			L.SetGlobal("shell", toLValue(L, shellMap))

			fn, err := L.LoadString(code)
			if err != nil {
				if mode == "keep-going" {
					rec.Error = &RecError{Stage: "lua-postmap", Message: err.Error()}
					results <- res{idx: idx, rec: rec, envE: &Error{Stage: "lua-postmap", Locator: rec.Locator, Message: err.Error()}}
					L.Close()
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("lua-postmap: %v", err)}
				L.Close()
				continue
			}
			L.Push(fn)
			done := make(chan struct{})
			var callErr error
			go func() {
				callErr = L.PCall(0, 1, nil)
				close(done)
			}()
			select {
			case <-done:
				if callErr != nil {
					if mode == "keep-going" {
						rec.Error = &RecError{Stage: "lua-postmap", Message: callErr.Error()}
						results <- res{idx: idx, rec: rec, envE: &Error{Stage: "lua-postmap", Locator: rec.Locator, Message: callErr.Error()}}
						L.Close()
						continue
					}
					results <- res{idx: idx, fatal: fmt.Errorf("lua-postmap: %v", callErr)}
					L.Close()
					continue
				}
			case <-time.After(200 * time.Millisecond):
				if mode == "keep-going" {
					rec.Error = &RecError{Stage: "lua-postmap", Message: "timeout"}
					results <- res{idx: idx, rec: rec, envE: &Error{Stage: "lua-postmap", Locator: rec.Locator, Message: "timeout"}}
					L.Close()
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("lua-postmap: timeout")}
				L.Close()
				continue
			}
			ret := L.Get(-1)
			L.Pop(1)
			rec.Post = fromLValue(ret)
			L.Close()
			results <- res{idx: idx, rec: rec}
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
