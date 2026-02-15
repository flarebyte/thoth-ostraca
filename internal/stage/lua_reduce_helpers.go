package stage

import (
	"fmt"
	"time"
)

// buildLuaReduceCode returns the Lua reduce code, wrapping expressions without explicit return.
func buildLuaReduceCode(in Envelope) string {
	code := in.Meta.Lua.ReduceInline
	if !containsReturn(code) {
		code = "return (" + code + ")"
	}
	return code
}

// reduceItemFromRecord converts a Record to the item value expected by the reducer.
func reduceItemFromRecord(rec Record) any {
	if rec.Post != nil {
		return rec.Post
	}
	m := map[string]any{"locator": rec.Locator}
	if rec.Meta != nil {
		m["meta"] = rec.Meta
	}
	if rec.Mapped != nil {
		m["mapped"] = rec.Mapped
	}
	if rec.Shell != nil {
		m["shell"] = map[string]any{"exitCode": rec.Shell.ExitCode, "stdout": rec.Shell.Stdout, "stderr": rec.Shell.Stderr}
	}
	if rec.Post != nil {
		m["post"] = rec.Post
	}
	return m
}

// runLuaReduce executes the reduce across all records and returns the final accumulator.
func runLuaReduce(in Envelope, code string) (any, error) {
	L := newMinimalLua()
	defer L.Close()
	var acc any
	for _, rec := range in.Records {
		item := reduceItemFromRecord(rec)
		L.SetGlobal("acc", toLValue(L, acc))
		L.SetGlobal("item", toLValue(L, item))
		fn, err := L.LoadString(code)
		if err != nil {
			return nil, fmt.Errorf("lua-reduce: %v", err)
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
				return nil, fmt.Errorf("lua-reduce: %v", callErr)
			}
		case <-time.After(200 * time.Millisecond):
			return nil, fmt.Errorf("lua-reduce: timeout")
		}
		ret := L.Get(-1)
		L.Pop(1)
		acc = fromLValue(ret)
	}
	return acc, nil
}
