package stage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaFilterRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine predicate
	pred := "return true"
	if in.Meta != nil && in.Meta.Lua != nil && in.Meta.Lua.FilterInline != "" {
		code := in.Meta.Lua.FilterInline
		// If code doesn't contain a return, wrap it
		if !containsReturn(code) {
			pred = "return (" + code + ")"
		} else {
			pred = code
		}
	}

	out := in
	out.Records = out.Records[:0]

	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	type res struct {
		idx   int
		keep  bool
		out   any
		envE  *Error
		fatal error
	}
	keeps := make([]bool, n)
	outs := make([]any, n)
	var envErrs []Error
	var mu sync.Mutex
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan res)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			var locator string
			var meta map[string]any
			var recErr *RecError
			switch rec := r.(type) {
			case Record:
				locator = rec.Locator
				meta = rec.Meta
				recErr = rec.Error
			case map[string]any:
				locVal, ok := rec["locator"]
				if !ok {
					if mode == "keep-going" {
						results <- res{idx: idx, keep: true, out: r, envE: &Error{Stage: "lua-filter", Message: "missing locator"}}
						continue
					}
					results <- res{idx: idx, fatal: errors.New("lua-filter: missing locator")}
					continue
				}
				s, ok := locVal.(string)
				if !ok {
					if mode == "keep-going" {
						results <- res{idx: idx, keep: true, out: r, envE: &Error{Stage: "lua-filter", Message: "invalid locator type"}}
						continue
					}
					results <- res{idx: idx, fatal: errors.New("lua-filter: invalid locator type")}
					continue
				}
				locator = s
				meta, _ = rec["meta"].(map[string]any)
			default:
				if mode == "keep-going" {
					results <- res{idx: idx, keep: true, out: r, envE: &Error{Stage: "lua-filter", Message: "invalid record type"}}
					continue
				}
				results <- res{idx: idx, fatal: errors.New("lua-filter: invalid record type")}
				continue
			}

			if recErr != nil {
				results <- res{idx: idx, keep: true, out: r}
				continue
			}

			L := lua.NewState(lua.Options{SkipOpenLibs: true})
			// Open minimal libs
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

			// Set globals
			L.SetGlobal("locator", lua.LString(locator))
			L.SetGlobal("meta", toLValue(L, meta))

			fn, err := L.LoadString(pred)
			if err != nil {
				if mode == "keep-going" {
					results <- res{idx: idx, keep: true, out: Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-filter", Message: err.Error()}}, envE: &Error{Stage: "lua-filter", Locator: locator, Message: err.Error()}}
					L.Close()
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("lua-filter: %v", err)}
				L.Close()
				continue
			}
			L.Push(fn)
			// Run with a simple timeout guard
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
						results <- res{idx: idx, keep: true, out: Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-filter", Message: callErr.Error()}}, envE: &Error{Stage: "lua-filter", Locator: locator, Message: callErr.Error()}}
						L.Close()
						continue
					}
					results <- res{idx: idx, fatal: fmt.Errorf("lua-filter: %v", callErr)}
					L.Close()
					continue
				}
			case <-time.After(200 * time.Millisecond):
				if mode == "keep-going" {
					results <- res{idx: idx, keep: true, out: Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-filter", Message: "timeout"}}, envE: &Error{Stage: "lua-filter", Locator: locator, Message: "timeout"}}
					L.Close()
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("lua-filter: timeout")}
				L.Close()
				continue
			}
			ret := L.Get(-1)
			L.Pop(1)
			keep := lua.LVAsBool(ret)
			L.Close()
			if keep {
				results <- res{idx: idx, keep: true, out: Record{Locator: locator, Meta: meta}}
			} else {
				results <- res{idx: idx, keep: false}
			}
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

func containsReturn(s string) bool {
	for i := 0; i+5 <= len(s); i++ {
		if s[i] == 'r' && i+6 <= len(s) && s[i:i+6] == "return" {
			return true
		}
	}
	return false
}

func toLValue(L *lua.LState, v any) lua.LValue {
	switch x := v.(type) {
	case nil:
		return lua.LNil
	case string:
		return lua.LString(x)
	case bool:
		if x {
			return lua.LTrue
		}
		return lua.LFalse
	case int:
		return lua.LNumber(float64(x))
	case int64:
		return lua.LNumber(float64(x))
	case float64:
		return lua.LNumber(x)
	case map[string]any:
		tbl := L.NewTable()
		for k, v2 := range x {
			tbl.RawSetString(k, toLValue(L, v2))
		}
		return tbl
	case []any:
		tbl := L.NewTable()
		for i, v2 := range x {
			tbl.RawSetInt(i+1, toLValue(L, v2))
		}
		return tbl
	default:
		return lua.LNil
	}
}

func init() { Register("lua-filter", luaFilterRunner) }
