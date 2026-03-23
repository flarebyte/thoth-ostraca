// File Guide for dev/ai agents:
// Purpose: Run the postMap stage that reshapes shell or mapped results into the final per-record structure used for output or persistence.
// Responsibilities:
// - Choose the default deterministic postMap behavior when no inline Lua is configured.
// - Execute configured postMap Lua over records in parallel.
// - Reuse the shared postMap parallel runner for consistent ordering and errors.
// Architecture notes:
// - The default no-Lua branch exists so pipeline actions can remain useful without forcing a script for simple shell workflows.
// - postMap is the bridge between analysis outputs and persistence because merge-meta later reads post.meta and post.nextMeta produced here.
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
