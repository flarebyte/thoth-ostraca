// File Guide for dev/ai agents:
// Purpose: Copy parsed Lua stage and Lua sandbox settings from config into runtime metadata.
// Responsibilities:
// - Apply filter, map, postMap, and reduce inline Lua code to the envelope metadata.
// - Apply sandbox limits and library toggles while preserving defaults for omitted fields.
// - Initialize Lua-related metadata sections only when the config actually uses them.
// Architecture notes:
// - Lua stage code and Lua sandbox settings are kept separate because script presence and sandbox tuning evolve independently.
// - Sandbox defaults are rehydrated here when the section exists so partial config blocks still get a complete runtime contract.
package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

func applyLuaMeta(out *Envelope, min config.Minimal) {
	if min.Filter.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.FilterInline = min.Filter.Inline
	}
	if min.Map.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.MapInline = min.Map.Inline
	}
	if min.PostMap.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.PostMapInline = min.PostMap.Inline
	}
	if min.Reduce.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.ReduceInline = min.Reduce.Inline
	}
}

func applyLuaSandboxMeta(out *Envelope, min config.Minimal) {
	if !min.LuaSandbox.HasSection {
		return
	}
	if out.Meta.LuaSandbox == nil {
		out.Meta.LuaSandbox = &LuaSandboxMeta{
			TimeoutMs:        defaultLuaTimeoutMs,
			InstructionLimit: defaultLuaInstructionLimit,
			MemoryLimitBytes: defaultLuaMemoryLimitBytes,
			Libs: LuaSandboxLibsMeta{
				Base:   true,
				Table:  true,
				String: true,
				Math:   true,
			},
			DeterministicRandom: true,
		}
	}
	if min.LuaSandbox.HasTimeoutMs {
		out.Meta.LuaSandbox.TimeoutMs = min.LuaSandbox.TimeoutMs
	}
	if min.LuaSandbox.HasInstructionLimit {
		out.Meta.LuaSandbox.InstructionLimit = min.LuaSandbox.InstructionLimit
	}
	if min.LuaSandbox.HasMemoryLimitBytes {
		out.Meta.LuaSandbox.MemoryLimitBytes = min.LuaSandbox.MemoryLimitBytes
	}
	if min.LuaSandbox.HasDeterministicRandom {
		out.Meta.LuaSandbox.DeterministicRandom = min.LuaSandbox.DeterministicRandom
	}
	if min.LuaSandbox.Libs.HasBase {
		out.Meta.LuaSandbox.Libs.Base = min.LuaSandbox.Libs.Base
	}
	if min.LuaSandbox.Libs.HasTable {
		out.Meta.LuaSandbox.Libs.Table = min.LuaSandbox.Libs.Table
	}
	if min.LuaSandbox.Libs.HasString {
		out.Meta.LuaSandbox.Libs.String = min.LuaSandbox.Libs.String
	}
	if min.LuaSandbox.Libs.HasMath {
		out.Meta.LuaSandbox.Libs.Math = min.LuaSandbox.Libs.Math
	}
}
