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
