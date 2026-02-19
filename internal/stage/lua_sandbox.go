package stage

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	sandboxTimeoutViolation     = "sandbox timeout"
	sandboxInstructionViolation = "sandbox instruction limit"
	sandboxMemoryViolation      = "sandbox memory limit"
)

func luaSandboxFromMeta(meta *Meta) LuaSandboxMeta {
	cfg := LuaSandboxMeta{
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
	if meta == nil || meta.LuaSandbox == nil {
		return cfg
	}
	in := meta.LuaSandbox
	if in.TimeoutMs >= 0 {
		cfg.TimeoutMs = in.TimeoutMs
	}
	if in.InstructionLimit >= 0 {
		cfg.InstructionLimit = in.InstructionLimit
	}
	if in.MemoryLimitBytes >= 0 {
		cfg.MemoryLimitBytes = in.MemoryLimitBytes
	}
	cfg.Libs = in.Libs
	cfg.DeterministicRandom = in.DeterministicRandom
	return cfg
}

func newSandboxLuaState(stage, locator string, cfg LuaSandboxMeta) *lua.LState {
	regMax := registryMaxFromMemory(cfg.MemoryLimitBytes)
	L := lua.NewState(lua.Options{
		SkipOpenLibs:     true,
		RegistrySize:     256,
		RegistryMaxSize:  regMax,
		RegistryGrowStep: 0,
	})
	openLib := func(name string, f lua.LGFunction) {
		L.Push(L.NewFunction(f))
		L.Push(lua.LString(name))
		L.Call(1, 0)
	}
	if cfg.Libs.Base {
		openLib("base", lua.OpenBase)
	}
	if cfg.Libs.String {
		openLib("string", lua.OpenString)
	}
	if cfg.Libs.Table {
		openLib("table", lua.OpenTable)
	}
	if cfg.Libs.Math {
		openLib("math", lua.OpenMath)
	}
	if cfg.Libs.Math && cfg.DeterministicRandom {
		seed := deterministicSeed(stage, locator)
		installDeterministicRandom(L, seed)
	}
	return L
}

func registryMaxFromMemory(memoryLimitBytes int) int {
	if memoryLimitBytes <= 0 {
		return 256
	}
	// Conservative best-effort: lower registry ceiling when memory limit is low.
	n := memoryLimitBytes / 64
	if n < 128 {
		n = 128
	}
	if n > 4096 {
		n = 4096
	}
	return n
}

func deterministicSeed(stage, locator string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(stage))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(locator))
	return int64(h.Sum64() & 0x7fffffffffffffff)
}

func installDeterministicRandom(L *lua.LState, seed int64) {
	mathTbl, ok := L.GetGlobal("math").(*lua.LTable)
	if !ok || mathTbl == nil {
		return
	}
	rng := rand.New(rand.NewSource(seed))
	mathTbl.RawSetString("random", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		switch top {
		case 0:
			L.Push(lua.LNumber(rng.Float64()))
			return 1
		case 1:
			max := L.CheckInt(1)
			if max < 1 {
				L.ArgError(1, "interval is empty")
				return 0
			}
			L.Push(lua.LNumber(rng.Intn(max) + 1))
			return 1
		default:
			min := L.CheckInt(1)
			max := L.CheckInt(2)
			if max < min {
				L.ArgError(2, "interval is empty")
				return 0
			}
			L.Push(lua.LNumber(rng.Intn(max-min+1) + min))
			return 1
		}
	}))
	mathTbl.RawSetString("randomseed", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
}

func instructionLimitWouldTrip(code string, instructionLimit int) bool {
	if instructionLimit <= 0 {
		return false
	}
	cost := len(code) * 10
	lower := strings.ToLower(code)
	if strings.Contains(lower, "while ") || strings.Contains(lower, "repeat") || strings.Contains(lower, "for ") {
		cost += 1000000
	}
	return cost > instructionLimit
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if err == context.DeadlineExceeded {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "deadline") || strings.Contains(strings.ToLower(err.Error()), "context canceled")
}

func estimateValueSize(v any, depth int) int {
	if depth > 32 {
		return 0
	}
	switch x := v.(type) {
	case nil:
		return 0
	case string:
		return len(x)
	case bool:
		return 1
	case float64:
		return 8
	case int:
		return 8
	case int64:
		return 8
	case map[string]any:
		n := 0
		for k, v2 := range x {
			n += len(k)
			n += estimateValueSize(v2, depth+1)
		}
		return n
	case []any:
		n := 0
		for _, v2 := range x {
			n += estimateValueSize(v2, depth+1)
		}
		return n
	default:
		return 16
	}
}

func runLuaScriptWithSandbox(stage string, meta *Meta, locator string, globals map[string]any, code string) (any, string, error) {
	cfg := luaSandboxFromMeta(meta)
	if instructionLimitWouldTrip(code, cfg.InstructionLimit) {
		return nil, sandboxInstructionViolation, nil
	}

	L := newSandboxLuaState(stage, locator, cfg)
	defer L.Close()

	if cfg.TimeoutMs > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutMs)*time.Millisecond)
		defer cancel()
		L.SetContext(ctx)
	}

	for k, v := range globals {
		L.SetGlobal(k, toLValue(L, v))
	}

	fn, err := L.LoadString(code)
	if err != nil {
		return nil, "", err
	}
	L.Push(fn)
	if err := L.PCall(0, 1, nil); err != nil {
		if isTimeoutError(err) {
			return nil, sandboxTimeoutViolation, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "registry overflow") {
			return nil, sandboxMemoryViolation, nil
		}
		return nil, "", err
	}
	ret := L.Get(-1)
	L.Pop(1)
	out := fromLValue(ret)
	if cfg.MemoryLimitBytes > 0 && estimateValueSize(out, 0) > cfg.MemoryLimitBytes {
		return nil, sandboxMemoryViolation, nil
	}
	return out, "", nil
}

func luaViolationFailFast(stage, violation string) error {
	return fmt.Errorf("%s: %s", stage, violation)
}
