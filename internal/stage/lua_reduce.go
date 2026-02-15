package stage

import (
	"context"
)

func luaReduceRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Default: count records
	if in.Meta == nil || in.Meta.Lua == nil || in.Meta.Lua.ReduceInline == "" {
		out := in
		out.Meta.Reduced = len(in.Records)
		return out, nil
	}
	code := buildLuaReduceCode(in)

	acc, err := runLuaReduce(in, code)
	if err != nil {
		return Envelope{}, err
	}

	out := in
	if out.Meta == nil {
		out.Meta = &Meta{}
	}
	out.Meta.Reduced = acc
	return out, nil
}

func init() { Register("lua-reduce", luaReduceRunner) }
