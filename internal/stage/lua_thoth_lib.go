package stage

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func installThothLib(L *lua.LState) {
	thoth := L.NewTable()
	thoth.RawSetString("ends_with", L.NewFunction(luaThothEndsWith))
	L.SetGlobal("thoth", thoth)
}

func luaThothEndsWith(L *lua.LState) int {
	s := L.CheckString(1)
	suffix := L.CheckString(2)
	L.Push(lua.LBool(strings.HasSuffix(s, suffix)))
	return 1
}
