package stage

import (
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func installThothLib(L *lua.LState) {
	L.SetGlobal("thoth", newThothLibTable(L))
}

func newThothLibTable(L *lua.LState) *lua.LTable {
	thoth := L.NewTable()
	thoth.RawSetString("all", L.NewFunction(luaThothAll))
	thoth.RawSetString("any", L.NewFunction(luaThothAny))
	thoth.RawSetString("contains", L.NewFunction(luaThothContains))
	thoth.RawSetString("ends_with", L.NewFunction(luaThothEndsWith))
	thoth.RawSetString("filter", L.NewFunction(luaThothFilter))
	thoth.RawSetString("find", L.NewFunction(luaThothFind))
	thoth.RawSetString("is_empty", L.NewFunction(luaThothIsEmpty))
	thoth.RawSetString("map", L.NewFunction(luaThothMap))
	thoth.RawSetString("push", L.NewFunction(luaThothPush))
	thoth.RawSetString("reduce", L.NewFunction(luaThothReduce))
	thoth.RawSetString("split", L.NewFunction(luaThothSplit))
	thoth.RawSetString("sort_keys", L.NewFunction(luaThothSortKeys))
	thoth.RawSetString("trim", L.NewFunction(luaThothTrim))
	return thoth
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

func luaThothFind(L *lua.LState) int {
	list := L.CheckTable(1)
	predicate := L.CheckFunction(2)
	var found lua.LValue = lua.LNil
	stop := false
	list.ForEach(func(_, item lua.LValue) {
		if stop {
			return
		}
		L.Push(predicate)
		L.Push(item)
		if err := L.PCall(1, 1, nil); err != nil {
			panic(err)
		}
		if lua.LVAsBool(L.Get(-1)) {
			found = item
			stop = true
		}
		L.Pop(1)
	})
	L.Push(found)
	return 1
}

func luaThothAny(L *lua.LState) int {
	list := L.CheckTable(1)
	predicate := L.CheckFunction(2)
	found := false
	stop := false
	list.ForEach(func(_, item lua.LValue) {
		if stop {
			return
		}
		L.Push(predicate)
		L.Push(item)
		if err := L.PCall(1, 1, nil); err != nil {
			panic(err)
		}
		if lua.LVAsBool(L.Get(-1)) {
			found = true
			stop = true
		}
		L.Pop(1)
	})
	L.Push(lua.LBool(found))
	return 1
}

func luaThothMap(L *lua.LState) int {
	list := L.CheckTable(1)
	fn := L.CheckFunction(2)
	out := L.NewTable()
	list.ForEach(func(_, item lua.LValue) {
		L.Push(fn)
		L.Push(item)
		if err := L.PCall(1, 1, nil); err != nil {
			panic(err)
		}
		out.Append(L.Get(-1))
		L.Pop(1)
	})
	L.Push(out)
	return 1
}

func luaThothFilter(L *lua.LState) int {
	list := L.CheckTable(1)
	fn := L.CheckFunction(2)
	out := L.NewTable()
	list.ForEach(func(_, item lua.LValue) {
		L.Push(fn)
		L.Push(item)
		if err := L.PCall(1, 1, nil); err != nil {
			panic(err)
		}
		if lua.LVAsBool(L.Get(-1)) {
			out.Append(item)
		}
		L.Pop(1)
	})
	L.Push(out)
	return 1
}

func luaThothReduce(L *lua.LState) int {
	list := L.CheckTable(1)
	acc := L.CheckAny(2)
	fn := L.CheckFunction(3)
	list.ForEach(func(_, item lua.LValue) {
		L.Push(fn)
		L.Push(acc)
		L.Push(item)
		if err := L.PCall(2, 1, nil); err != nil {
			panic(err)
		}
		acc = L.Get(-1)
		L.Pop(1)
	})
	L.Push(acc)
	return 1
}

func luaThothAll(L *lua.LState) int {
	list := L.CheckTable(1)
	predicate := L.CheckFunction(2)
	all := true
	stop := false
	list.ForEach(func(_, item lua.LValue) {
		if stop {
			return
		}
		L.Push(predicate)
		L.Push(item)
		if err := L.PCall(1, 1, nil); err != nil {
			panic(err)
		}
		if !lua.LVAsBool(L.Get(-1)) {
			all = false
			stop = true
		}
		L.Pop(1)
	})
	L.Push(lua.LBool(all))
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
