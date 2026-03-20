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
	thoth.RawSetString("is_empty", L.NewFunction(luaThothIsEmpty))
	thoth.RawSetString("push", L.NewFunction(luaThothPush))
	thoth.RawSetString("split", L.NewFunction(luaThothSplit))
	thoth.RawSetString("sort_keys", L.NewFunction(luaThothSortKeys))
	thoth.RawSetString("trim", L.NewFunction(luaThothTrim))
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

func luaThothSplit(L *lua.LState) int {
	s := L.CheckString(1)
	sep := L.CheckString(2)
	out := L.NewTable()
	if sep == "" {
		for _, part := range strings.Split(s, sep) {
			out.Append(lua.LString(part))
		}
		L.Push(out)
		return 1
	}
	for _, part := range strings.Split(s, sep) {
		out.Append(lua.LString(part))
	}
	L.Push(out)
	return 1
}

func luaThothTrim(L *lua.LState) int {
	s := L.CheckString(1)
	L.Push(lua.LString(strings.TrimSpace(s)))
	return 1
}

func luaThothIsEmpty(L *lua.LState) int {
	tbl := L.CheckTable(1)
	empty := true
	tbl.ForEach(func(_, _ lua.LValue) {
		empty = false
	})
	L.Push(lua.LBool(empty))
	return 1
}
