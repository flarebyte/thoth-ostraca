package stage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine transform
	code := "return { locator = locator, meta = meta }"
	if in.Meta != nil && in.Meta.Lua != nil && in.Meta.Lua.MapInline != "" {
		// Allow expressions without explicit return
		c := in.Meta.Lua.MapInline
		if !containsReturn(c) {
			code = "return (" + c + ")"
		} else {
			code = c
		}
	}

	out := in
	mode, _ := errorMode(in.Meta)
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
	var envErrs []Error
	var mu sync.Mutex
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			var locator string
			var meta map[string]any
			if rec, ok := r.(Record); ok {
				if rec.Error != nil {
					results <- res{idx: idx, rec: r}
					continue
				}
				locator = rec.Locator
				meta = rec.Meta
			} else if m, ok := r.(map[string]any); ok {
				locVal, ok := m["locator"]
				if !ok {
					if mode == "keep-going" {
						results <- res{idx: idx, rec: r, envE: &Error{Stage: "lua-map", Message: "missing locator"}}
						continue
					}
					results <- res{idx: idx, fatal: errors.New("lua-map: missing locator")}
					continue
				}
				s, ok := locVal.(string)
				if !ok {
					if mode == "keep-going" {
						results <- res{idx: idx, rec: r, envE: &Error{Stage: "lua-map", Message: "invalid locator type"}}
						continue
					}
					results <- res{idx: idx, fatal: errors.New("lua-map: invalid locator type")}
					continue
				}
				locator = s
				meta, _ = m["meta"].(map[string]any)
			} else {
				if mode == "keep-going" {
					results <- res{idx: idx, rec: r, envE: &Error{Stage: "lua-map", Message: "invalid record type"}}
					continue
				}
				results <- res{idx: idx, fatal: errors.New("lua-map: invalid record type")}
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

			L.SetGlobal("locator", lua.LString(locator))
			L.SetGlobal("meta", toLValue(L, meta))

			fn, err := L.LoadString(code)
			if err != nil {
				if mode == "keep-going" {
					results <- res{idx: idx, rec: Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-map", Message: err.Error()}}, envE: &Error{Stage: "lua-map", Locator: locator, Message: err.Error()}}
					L.Close()
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("lua-map: %v", err)}
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
						results <- res{idx: idx, rec: Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-map", Message: callErr.Error()}}, envE: &Error{Stage: "lua-map", Locator: locator, Message: callErr.Error()}}
						L.Close()
						continue
					}
					results <- res{idx: idx, fatal: fmt.Errorf("lua-map: %v", callErr)}
					L.Close()
					continue
				}
			case <-time.After(200 * time.Millisecond):
				if mode == "keep-going" {
					results <- res{idx: idx, rec: Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-map", Message: "timeout"}}, envE: &Error{Stage: "lua-map", Locator: locator, Message: "timeout"}}
					L.Close()
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("lua-map: timeout")}
				L.Close()
				continue
			}
			ret := L.Get(-1)
			L.Pop(1)
			mapped := fromLValue(ret)
			L.Close()
			results <- res{idx: idx, rec: Record{Locator: locator, Meta: meta, Mapped: mapped}}
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

func fromLValue(v lua.LValue) any {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(v)
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTString:
		return v.String()
	case lua.LTTable:
		t := v.(*lua.LTable)
		// Decide object vs array by checking numeric keys 1..n
		// We'll build a map if non-sequential keys exist
		// First try array
		arr := []any{}
		isArray := true
		t.ForEach(func(k, val lua.LValue) {
			if isArray {
				if lk, ok := k.(lua.LNumber); ok && int(lk) == len(arr)+1 {
					arr = append(arr, fromLValue(val))
				} else {
					isArray = false
				}
			}
		})
		if isArray {
			return arr
		}
		obj := map[string]any{}
		t.ForEach(func(k, val lua.LValue) {
			obj[k.String()] = fromLValue(val)
		})
		return obj
	default:
		return nil
	}
}

func init() { Register("lua-map", luaMapRunner) }
