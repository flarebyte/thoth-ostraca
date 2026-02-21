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
