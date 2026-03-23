// File Guide for dev/ai agents:
// Purpose: Run the final Lua reduce stage that collapses per-record results into one aggregate value.
// Responsibilities:
// - Provide the default aggregate behavior when no reduce script is configured.
// - Execute the configured reduce Lua across the current record set.
// - Store the aggregate result in meta.reduced for final output.
// Architecture notes:
// - Reduced output lives in meta.reduced by contract so records remain available alongside the aggregate summary.
// - The default reduction is intentionally simple record counting to keep the stage predictable even when configured but unused.
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
