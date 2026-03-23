// File Guide for dev/ai agents:
// Purpose: Hold shared validate-config defaults and the top-level function that projects parsed config into runtime metadata.
// Responsibilities:
// - Define runtime default constants used by config application helpers.
// - Deep-copy generic config values before attaching them to the envelope.
// - Coordinate the concern-specific apply* helpers into one metadata projection pass.
// Architecture notes:
// - Default constants live here so all apply helpers share one source of truth instead of embedding fallback values in multiple files.
// - applyMinimalToMeta is intentionally orchestration-only; detailed field logic is delegated to smaller helper files to keep config growth manageable.
package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

const defaultMaxRecordsInMemory = 10000
const defaultLuaTimeoutMs = 2000
const defaultLuaInstructionLimit = 1000000
const defaultLuaMemoryLimitBytes = 8388608
const defaultShellProgram = "bash"
const defaultShellWorkingDir = "."
const defaultShellTimeoutMs = 60000
const defaultShellCaptureMaxBytes = 1048576
const defaultShellTermGraceMs = 2000
const defaultUIProgressIntervalMs = 500

func deepCopyAny(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[k] = deepCopyAny(vv)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i := range x {
			out[i] = deepCopyAny(x[i])
		}
		return out
	default:
		return x
	}
}

// applyMinimalToMeta mutates out.Meta to reflect values from the parsed minimal config.
// It mirrors the original field population and presence checks exactly.
func applyMinimalToMeta(out *Envelope, min config.Minimal) {
	if out.Meta == nil {
		out.Meta = &Meta{}
	}
	out.Meta.Config = &ConfigMeta{ConfigVersion: min.ConfigVersion, Action: min.Action}
	out.Meta.ConfigPath = "" // do not persist configPath in output

	applyDiscoveryMeta(out, min)
	applyValidationMeta(out, min)
	applyLimitsMeta(out, min)
	applyLuaMeta(out, min)
	applyLuaSandboxMeta(out, min)
	applyShellMeta(out, min)
	applyOutputMeta(out, min)
	applyPersistMeta(out, min)
	applyUpdateMeta(out, min)
	applyDiffMeta(out, min)
	applyErrorsMeta(out, min)
	applyFileInfoMeta(out, min)
	applyGitMeta(out, min)
	applyWorkersMeta(out, min)
	applyUIMeta(out, min)
	applyLocatorPolicyMeta(out, min)
}
