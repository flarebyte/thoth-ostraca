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

func TestInstallThothLib(t *testing.T) {
	t.Parallel()

	L := lua.NewState()
	defer L.Close()

	installThothLib(L)

	thoth, ok := L.GetGlobal("thoth").(*lua.LTable)
	if !ok || thoth == nil {
		t.Fatalf("expected global thoth table, got %T", L.GetGlobal("thoth"))
	}
	if thoth.RawGetString("ends_with").Type() != lua.LTFunction {
		t.Fatalf("expected thoth.ends_with function to be registered")
	}
}
