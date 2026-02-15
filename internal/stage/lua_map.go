package stage

import (
	"context"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func luaMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine transform
	code := buildLuaMapCode(in)

	out := in
	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan luaMapRes)
	var wg sync.WaitGroup
	var envErrs []Error
	var mu sync.Mutex
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			rec, envE, fatal := processLuaMapRecord(r, code, mode)
			results <- luaMapRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
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
