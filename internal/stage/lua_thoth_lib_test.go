package stage

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

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

func TestInstallThothLib(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	installThothLib(L)

	thoth, ok := L.GetGlobal("thoth").(*lua.LTable)
	if !ok || thoth == nil {
		t.Fatalf("expected global thoth table, got %T", L.GetGlobal("thoth"))
	}
	if thoth.RawGetString("contains").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.contains function to be registered")
	}
	if thoth.RawGetString("ends_with").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.ends_with function to be registered")
	}
	if thoth.RawGetString("push").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.push function to be registered")
	}
	if thoth.RawGetString("sort_keys").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.sort_keys function to be registered")
	}
}
