package stage

import (
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func installThothLib(L *lua.LState) {
	thoth := L.NewTable()
	thoth.RawSetString("contains", L.NewFunction(luaThothContains))
	thoth.RawSetString("ends_with", L.NewFunction(luaThothEndsWith))
	thoth.RawSetString("push", L.NewFunction(luaThothPush))
	thoth.RawSetString("sort_keys", L.NewFunction(luaThothSortKeys))
	L.SetGlobal("thoth", thoth)
}

func luaThothEndsWith(L *lua.LState) int {
	s := L.CheckString(1)
	suffix := L.CheckString(2)
	L.Push(lua.LBool(strings.HasSuffix(s, suffix)))
	return 1
}

func luaThothSortKeys(L *lua.LState) int {
	tbl := L.CheckTable(1)
	keys := make([]string, 0)
	tbl.ForEach(func(key, _ lua.LValue) {
		if key.Type() == lua.LTString {
			keys = append(keys, key.String())
		}
	})
	sort.Strings(keys)
	out := L.NewTable()
	for _, key := range keys {
		out.Append(lua.LString(key))
	}
	L.Push(out)
	return 1
}

func luaThothContains(L *lua.LState) int {
	list := L.CheckTable(1)
	value := L.CheckAny(2)
	found := false
	list.ForEach(func(_, item lua.LValue) {
		if found {
			return
		}
		if L.RawEqual(item, value) {
			found = true
		}
	})
	L.Push(lua.LBool(found))
	return 1
}

func luaThothPush(L *lua.LState) int {
	list := L.CheckTable(1)
	value := L.CheckAny(2)
	list.Append(value)
	L.Push(list)
	return 1
}
