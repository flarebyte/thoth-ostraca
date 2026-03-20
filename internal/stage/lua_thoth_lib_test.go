package stage

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newLuaStateWithThothLib(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	L.SetGlobal("thoth", newThothLibTable(L))
	return L
}

func TestLuaThothEndsWith(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		s      string
		suffix string
		want   bool
	}{
		{
			name:   "match suffix",
			s:      "internal/metafile/write.go",
			suffix: ".go",
			want:   true,
		},
		{
			name:   "reject non matching suffix",
			s:      "internal/metafile/write.go",
			suffix: "_test.go",
			want:   false,
		},
		{
			name:   "empty suffix matches",
			s:      "write.go",
			suffix: "",
			want:   true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			L := lua.NewState()
			defer L.Close()

			L.Push(lua.LString(tc.s))
			L.Push(lua.LString(tc.suffix))

			gotN := luaThothEndsWith(L)
			if gotN != 1 {
				t.Fatalf("expected 1 return value, got %d", gotN)
			}

			got, ok := L.Get(-1).(lua.LBool)
			if !ok {
				t.Fatalf("expected boolean result, got %T", L.Get(-1))
			}
			if bool(got) != tc.want {
				t.Fatalf(
					"luaThothEndsWith(%q, %q) = %v, want %v",
					tc.s,
					tc.suffix,
					bool(got),
					tc.want,
				)
			}
		})
	}
}

func TestLuaThothCopy(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local original = {alpha = "beta"}
    result = thoth.copy(original)
    original.alpha = "changed"
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	gotTbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.GetGlobal("result"))
	}
	if gotTbl.RawGetString("alpha").String() != "beta" {
		t.Fatalf(
			"copy.alpha = %q, want %q",
			gotTbl.RawGetString("alpha").String(),
			"beta",
		)
	}
}

func TestLuaThothDeepCopy(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local original = {nested = {alpha = "beta"}}
    result = thoth.deep_copy(original)
    original.nested.alpha = "changed"
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	gotTbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.GetGlobal("result"))
	}
	nested, ok := gotTbl.RawGetString("nested").(*lua.LTable)
	if !ok {
		t.Fatalf("expected nested table, got %T", gotTbl.RawGetString("nested"))
	}
	if nested.RawGetString("alpha").String() != "beta" {
		t.Fatalf(
			"deep_copy.nested.alpha = %q, want %q",
			nested.RawGetString("alpha").String(),
			"beta",
		)
	}
}

func TestLuaThothSortKeys(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("beta", lua.LNumber(2))
	tbl.RawSetString("alpha", lua.LNumber(1))
	tbl.RawSetString("gamma", lua.LNumber(3))

	L.Push(tbl)

	gotN := luaThothSortKeys(L)
	if gotN != 1 {
		t.Fatalf("expected 1 return value, got %d", gotN)
	}

	gotTbl, ok := L.Get(-1).(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.Get(-1))
	}
	got := []string{
		gotTbl.RawGetInt(1).String(),
		gotTbl.RawGetInt(2).String(),
		gotTbl.RawGetInt(3).String(),
	}
	want := []string{"alpha", "beta", "gamma"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sort_keys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLuaThothContains(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value lua.LValue
		want  bool
	}{
		{
			name:  "contains string",
			value: lua.LString("beta"),
			want:  true,
		},
		{
			name:  "does not contain string",
			value: lua.LString("delta"),
			want:  false,
		},
		{
			name:  "type must match",
			value: lua.LNumber(2),
			want:  false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			L := lua.NewState()
			defer L.Close()

			list := L.NewTable()
			list.Append(lua.LString("alpha"))
			list.Append(lua.LString("beta"))
			list.Append(lua.LString("gamma"))

			L.Push(list)
			L.Push(tc.value)

			gotN := luaThothContains(L)
			if gotN != 1 {
				t.Fatalf("expected 1 return value, got %d", gotN)
			}

			got, ok := L.Get(-1).(lua.LBool)
			if !ok {
				t.Fatalf("expected boolean result, got %T", L.Get(-1))
			}
			if bool(got) != tc.want {
				t.Fatalf("contains = %v, want %v", bool(got), tc.want)
			}
		})
	}
}

func TestLuaThothPush(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	list := L.NewTable()
	list.Append(lua.LString("alpha"))

	L.Push(list)
	L.Push(lua.LString("beta"))

	gotN := luaThothPush(L)
	if gotN != 1 {
		t.Fatalf("expected 1 return value, got %d", gotN)
	}

	gotTbl, ok := L.Get(-1).(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.Get(-1))
	}
	if gotTbl.RawGetInt(1).String() != "alpha" {
		t.Fatalf("expected first element alpha, got %q", gotTbl.RawGetInt(1))
	}
	if gotTbl.RawGetInt(2).String() != "beta" {
		t.Fatalf("expected second element beta, got %q", gotTbl.RawGetInt(2))
	}
}

func TestLuaThothFlatten(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"alpha", {"beta", {"gamma"}}, "delta"}
    result = thoth.flatten(items)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	gotTbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.GetGlobal("result"))
	}
	want := []string{"alpha", "beta", "gamma", "delta"}
	for i, exp := range want {
		if gotTbl.RawGetInt(i+1).String() != exp {
			t.Fatalf(
				"flatten[%d] = %q, want %q",
				i,
				gotTbl.RawGetInt(i+1).String(),
				exp,
			)
		}
	}
}

func TestLuaThothSplit(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LString("alpha,beta,gamma"))
	L.Push(lua.LString(","))

	gotN := luaThothSplit(L)
	if gotN != 1 {
		t.Fatalf("expected 1 return value, got %d", gotN)
	}

	gotTbl, ok := L.Get(-1).(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.Get(-1))
	}
	want := []string{"alpha", "beta", "gamma"}
	for i, exp := range want {
		if gotTbl.RawGetInt(i+1).String() != exp {
			t.Fatalf(
				"split[%d] = %q, want %q",
				i,
				gotTbl.RawGetInt(i+1).String(),
				exp,
			)
		}
	}
}

func TestLuaThothSortValues(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.Append(lua.LString("beta"))
	tbl.Append(lua.LString("alpha"))
	tbl.Append(lua.LString("gamma"))

	L.Push(tbl)

	gotN := luaThothSortValues(L)
	if gotN != 1 {
		t.Fatalf("expected 1 return value, got %d", gotN)
	}

	gotTbl, ok := L.Get(-1).(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.Get(-1))
	}
	want := []string{"alpha", "beta", "gamma"}
	for i, exp := range want {
		if gotTbl.RawGetInt(i+1).String() != exp {
			t.Fatalf(
				"sort_values[%d] = %q, want %q",
				i,
				gotTbl.RawGetInt(i+1).String(),
				exp,
			)
		}
	}
}

func TestLuaThothTrim(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LString("  alpha beta  \n"))

	gotN := luaThothTrim(L)
	if gotN != 1 {
		t.Fatalf("expected 1 return value, got %d", gotN)
	}

	got, ok := L.Get(-1).(lua.LString)
	if !ok {
		t.Fatalf("expected string result, got %T", L.Get(-1))
	}
	if string(got) != "alpha beta" {
		t.Fatalf("trim = %q, want %q", string(got), "alpha beta")
	}
}

func TestLuaThothIsEmpty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fill func(*lua.LState) *lua.LTable
		want bool
	}{
		{
			name: "empty table",
			fill: func(L *lua.LState) *lua.LTable {
				return L.NewTable()
			},
			want: true,
		},
		{
			name: "array table",
			fill: func(L *lua.LState) *lua.LTable {
				tbl := L.NewTable()
				tbl.Append(lua.LString("alpha"))
				return tbl
			},
			want: false,
		},
		{
			name: "map table",
			fill: func(L *lua.LState) *lua.LTable {
				tbl := L.NewTable()
				tbl.RawSetString("alpha", lua.LString("beta"))
				return tbl
			},
			want: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			L := lua.NewState()
			defer L.Close()

			L.Push(tc.fill(L))

			gotN := luaThothIsEmpty(L)
			if gotN != 1 {
				t.Fatalf("expected 1 return value, got %d", gotN)
			}

			got, ok := L.Get(-1).(lua.LBool)
			if !ok {
				t.Fatalf("expected boolean result, got %T", L.Get(-1))
			}
			if bool(got) != tc.want {
				t.Fatalf("is_empty = %v, want %v", bool(got), tc.want)
			}
		})
	}
}

func TestLuaThothFind(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"alpha", "beta", "gamma"}
    result = thoth.find(items, function(item)
      return item == "beta"
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	got, ok := L.GetGlobal("result").(lua.LString)
	if !ok {
		t.Fatalf("expected string result, got %T", L.GetGlobal("result"))
	}
	if string(got) != "beta" {
		t.Fatalf("find = %q, want %q", string(got), "beta")
	}
}

func TestLuaThothFindNotFound(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"alpha", "beta", "gamma"}
    result = thoth.find(items, function(item)
      return item == "delta"
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	if L.GetGlobal("result") != lua.LNil {
		t.Fatalf("expected nil result, got %T", L.GetGlobal("result"))
	}
}

func TestLuaThothAny(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"alpha", "beta", "gamma"}
    result = thoth.any(items, function(item)
      return item == "beta"
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	got, ok := L.GetGlobal("result").(lua.LBool)
	if !ok {
		t.Fatalf("expected bool result, got %T", L.GetGlobal("result"))
	}
	if !bool(got) {
		t.Fatalf("any = false, want true")
	}
}

func TestLuaThothAll(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"alpha", "beta", "gamma"}
    result = thoth.all(items, function(item)
      return string.len(item) >= 4
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	got, ok := L.GetGlobal("result").(lua.LBool)
	if !ok {
		t.Fatalf("expected bool result, got %T", L.GetGlobal("result"))
	}
	if !bool(got) {
		t.Fatalf("all = false, want true")
	}
}

func TestLuaThothAllFalse(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"alpha", "beta", "gamma"}
    result = thoth.all(items, function(item)
      return item == "beta"
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	got, ok := L.GetGlobal("result").(lua.LBool)
	if !ok {
		t.Fatalf("expected bool result, got %T", L.GetGlobal("result"))
	}
	if bool(got) {
		t.Fatalf("all = true, want false")
	}
}

func TestLuaThothMap(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"go", "lua", "cue"}
    result = thoth.map(items, function(item)
      return string.upper(item)
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	gotTbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.GetGlobal("result"))
	}
	want := []string{"GO", "LUA", "CUE"}
	for i, exp := range want {
		if gotTbl.RawGetInt(i+1).String() != exp {
			t.Fatalf(
				"map[%d] = %q, want %q",
				i,
				gotTbl.RawGetInt(i+1).String(),
				exp,
			)
		}
	}
}

func TestLuaThothFilter(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {"write.go", "write_test.go", "read.go"}
    result = thoth.filter(items, function(item)
      return not thoth.ends_with(item, "_test.go")
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	gotTbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("expected table result, got %T", L.GetGlobal("result"))
	}
	want := []string{"write.go", "read.go"}
	for i, exp := range want {
		if gotTbl.RawGetInt(i+1).String() != exp {
			t.Fatalf(
				"filter[%d] = %q, want %q",
				i,
				gotTbl.RawGetInt(i+1).String(),
				exp,
			)
		}
	}
}

func TestLuaThothReduce(t *testing.T) {
	t.Parallel()

	L := newLuaStateWithThothLib(t)
	defer L.Close()

	if err := L.DoString(`
    local items = {10, 20, 30}
    result = thoth.reduce(items, 0, function(acc, item)
      return acc + item
    end)
  `); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}

	got, ok := L.GetGlobal("result").(lua.LNumber)
	if !ok {
		t.Fatalf("expected numeric result, got %T", L.GetGlobal("result"))
	}
	if float64(got) != 60 {
		t.Fatalf("reduce = %v, want 60", float64(got))
	}
}

func TestInstallThothLib(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	installThothLib(L)

	thoth, ok := L.GetGlobal("thoth").(*lua.LTable)
	if !ok || thoth == nil {
		t.Fatalf("expected global thoth table, got %T", L.GetGlobal("thoth"))
	}
	if thoth.RawGetString("all").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.all function to be registered")
	}
	if thoth.RawGetString("any").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.any function to be registered")
	}
	if thoth.RawGetString("contains").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.contains function to be registered")
	}
	if thoth.RawGetString("copy").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.copy function to be registered")
	}
	if thoth.RawGetString("deep_copy").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.deep_copy function to be registered")
	}
	if thoth.RawGetString("ends_with").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.ends_with function to be registered")
	}
	if thoth.RawGetString("filter").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.filter function to be registered")
	}
	if thoth.RawGetString("find").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.find function to be registered")
	}
	if thoth.RawGetString("flatten").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.flatten function to be registered")
	}
	if thoth.RawGetString("is_empty").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.is_empty function to be registered")
	}
	if thoth.RawGetString("map").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.map function to be registered")
	}
	if thoth.RawGetString("push").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.push function to be registered")
	}
	if thoth.RawGetString("reduce").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.reduce function to be registered")
	}
	if thoth.RawGetString("split").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.split function to be registered")
	}
	if thoth.RawGetString("sort_keys").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.sort_keys function to be registered")
	}
	if thoth.RawGetString("sort_values").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.sort_values function to be registered")
	}
	if thoth.RawGetString("trim").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.trim function to be registered")
	}
}
