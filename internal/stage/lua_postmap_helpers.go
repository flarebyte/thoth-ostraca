package stage

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const luaPostMapStage = "lua-postmap"

type luaPostMapRes struct {
	idx   int
	rec   Record
	envE  *Error
	fatal error
}

// buildLuaPostMapCode returns the Lua code to execute for postmap.
// Caller decides when to use the default non-Lua path.
func buildLuaPostMapCode(in Envelope) string {
	code := in.Meta.Lua.PostMapInline
	if !containsReturn(code) {
		code = "return (" + code + ")"
	}
	return code
}

// processDefaultPostMapRecord applies the deterministic default postmap when no inline code is provided.
func processDefaultPostMapRecord(rec Record, mode string) (Record, *Error, error) {
	if rec.Error != nil {
		return rec, nil, nil
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
	return rec, nil, nil
}

// processLuaPostMapRecord runs the Lua postmap code against a record.
func processLuaPostMapRecord(rec Record, code string, mode string) (Record, *Error, error) {
	L := newMinimalLua()
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
			rec.Error = &RecError{Stage: luaPostMapStage, Message: err.Error()}
			L.Close()
			return rec, &Error{Stage: luaPostMapStage, Locator: rec.Locator, Message: err.Error()}, nil
		}
		L.Close()
		return Record{}, nil, fmt.Errorf("lua-postmap: %v", err)
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
				rec.Error = &RecError{Stage: luaPostMapStage, Message: callErr.Error()}
				L.Close()
				return rec, &Error{Stage: luaPostMapStage, Locator: rec.Locator, Message: callErr.Error()}, nil
			}
			L.Close()
			return Record{}, nil, fmt.Errorf("lua-postmap: %v", callErr)
		}
	case <-time.After(200 * time.Millisecond):
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: luaPostMapStage, Message: "timeout"}
			L.Close()
			return rec, &Error{Stage: luaPostMapStage, Locator: rec.Locator, Message: "timeout"}, nil
		}
		L.Close()
		return Record{}, nil, fmt.Errorf("lua-postmap: timeout")
	}
	ret := L.Get(-1)
	L.Pop(1)
	rec.Post = fromLValue(ret)
	L.Close()
	return rec, nil, nil
}

// runPostMapParallel runs the provided per-record postmap function across
// records in parallel, collecting envelope errors and first fatal error.
func runPostMapParallel(in Envelope, mode string, fn func(Record) (Record, *Error, error)) (Envelope, error) {
	out := in
	n := len(in.Records)
	workers := getWorkers(in.Meta)
	results := runIndexedParallel(n, workers, func(idx int) luaPostMapRes {
		r := in.Records[idx]
		rec, envE, fatal := fn(r)
		return luaPostMapRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
	})
	var firstErr error
	for _, rr := range results {
		if rr.envE != nil {
			out.Errors = append(out.Errors, *rr.envE)
		}
		if rr.fatal != nil && firstErr == nil {
			firstErr = rr.fatal
		}
		out.Records[rr.idx] = rr.rec
	}
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	return out, nil
}
