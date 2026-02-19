package stage

import (
	"fmt"
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
func processLuaPostMapRecord(rec Record, code string, mode string, metaCfg *Meta) (Record, *Error, error) {
	var shellMap map[string]any
	if rec.Shell != nil {
		shellMap = map[string]any{
			"exitCode": rec.Shell.ExitCode,
			"timedOut": rec.Shell.TimedOut,
		}
		if rec.Shell.Stdout != nil {
			shellMap["stdout"] = *rec.Shell.Stdout
		}
		if rec.Shell.Stderr != nil {
			shellMap["stderr"] = *rec.Shell.Stderr
		}
		if rec.Shell.Error != nil {
			shellMap["error"] = *rec.Shell.Error
		}
	}
	ret, violation, err := runLuaScriptWithSandbox(luaPostMapStage, metaCfg, rec.Locator, map[string]any{
		"locator": rec.Locator,
		"meta":    rec.Meta,
		"mapped":  rec.Mapped,
		"shell":   shellMap,
	}, code)
	if err != nil {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: luaPostMapStage, Message: err.Error()}
			return rec, &Error{Stage: luaPostMapStage, Locator: rec.Locator, Message: err.Error()}, nil
		}
		return Record{}, nil, fmt.Errorf("lua-postmap: %v", err)
	}
	if violation != "" {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: luaPostMapStage, Message: violation}
			return rec, &Error{Stage: luaPostMapStage, Locator: rec.Locator, Message: violation}, nil
		}
		return Record{}, nil, luaViolationFailFast(luaPostMapStage, violation)
	}
	rec.Post = ret
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
