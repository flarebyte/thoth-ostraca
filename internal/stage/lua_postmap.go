package stage

import (
	"context"
)

func luaPostMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	mode, _ := errorMode(in.Meta)
	// If no inline, apply deterministic default without Lua to keep ordering stable
	if in.Meta == nil || in.Meta.Lua == nil || in.Meta.Lua.PostMapInline == "" {
		return runPostMapParallel(in, mode, func(r Record) (Record, *Error, error) {
			return processDefaultPostMapRecord(r, mode)
		})
	}

	code := buildLuaPostMapCode(in)
	return runPostMapParallel(in, mode, func(r Record) (Record, *Error, error) {
		return processLuaPostMapRecord(r, code, mode, in.Meta)
	})
}

func init() { Register("lua-postmap", luaPostMapRunner) }
