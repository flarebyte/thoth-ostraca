// File Guide for dev/ai agents:
// Purpose: Execute the tiny expectedLua contract used by update-meta and diff-meta to derive per-locator expected metadata.
// Responsibilities:
// - Wrap inline expectedLua snippets into the runtime program shape the sandbox expects.
// - Run the snippet with locator and existingMeta context.
// - Validate that the snippet returns a function and then an object result.
// Architecture notes:
// - expectedLua is wrapped to require a returned function on purpose so config authors write explicit per-locator transforms instead of implicit top-level scripts.
// - The type-marker key is an internal sentinel used only to distinguish a non-function return from a valid object result.
package stage

import "fmt"

const expectedLuaTypeMarkerKey = "__thoth_expected_lua_error"

func buildExpectedLuaProgram(inline string) string {
	return "local __thoth_expected_fn = (function()\n" + inline + "\nend)()\n" +
		"if type(__thoth_expected_fn) ~= \"function\" then\n" +
		"  return { [\"" + expectedLuaTypeMarkerKey + "\"] = \"non-function\" }\n" +
		"end\n" +
		"return __thoth_expected_fn(locator, existingMeta)\n"
}

func runExpectedLuaInline(stageName string, metaCfg *Meta, locator string, existing map[string]any, inline string) (map[string]any, string, error) {
	code := buildExpectedLuaProgram(inline)
	ret, violation, err := runLuaScriptWithSandbox(stageName, metaCfg, locator, map[string]any{
		"locator":      locator,
		"existingMeta": existing,
	}, code)
	if err != nil || violation != "" {
		return nil, violation, err
	}
	rm, ok := asStringMap(ret)
	if !ok {
		return nil, "", fmt.Errorf("expectedLua must return object")
	}
	if marker, ok := rm[expectedLuaTypeMarkerKey].(string); ok && marker == "non-function" && len(rm) == 1 {
		return nil, "", fmt.Errorf("expectedLua must return function")
	}
	if cp, ok := deepCopyAny(rm).(map[string]any); ok {
		return cp, "", nil
	}
	return map[string]any{}, "", nil
}
