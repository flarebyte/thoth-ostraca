package stage

import (
	"context"

	lua "github.com/yuin/gopher-lua"
)

func luaMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine transform
	code := buildLuaMapCode(in)

	out := in
	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	workers := getWorkers(in.Meta)
	results := runIndexedParallel(n, workers, func(idx int) recordParallelRes {
		r := in.Records[idx]
		rec, envE, fatal := processLuaMapRecord(r, code, mode, in.Meta)
		return recordParallelRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
	})
	return mergeRecordParallelResults(out, results)
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
