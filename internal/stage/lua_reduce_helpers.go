package stage

import (
	"fmt"
)

const luaReduceStage = "lua-reduce"

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
	var acc any
	for _, rec := range in.Records {
		item := reduceItemFromRecord(rec)
		locator := rec.Locator
		if locator == "" {
			locator = "reduce"
		}
		ret, violation, err := runLuaScriptWithSandbox(luaReduceStage, in.Meta, locator, map[string]any{
			"acc":  acc,
			"item": item,
		}, code)
		if err != nil {
			return nil, fmt.Errorf("lua-reduce: %v", err)
		}
		if violation != "" {
			return nil, luaViolationFailFast(luaReduceStage, violation)
		}
		acc = ret
	}
	return acc, nil
}
