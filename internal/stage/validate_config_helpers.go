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
	applyUpdateMeta(out, min)
	applyDiffMeta(out, min)
	applyErrorsMeta(out, min)
	applyFileInfoMeta(out, min)
	applyGitMeta(out, min)
	applyWorkersMeta(out, min)
	applyUIMeta(out, min)
	applyLocatorPolicyMeta(out, min)
}
