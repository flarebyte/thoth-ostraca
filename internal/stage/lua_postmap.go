package stage

import (
	"context"
	"errors"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaPostMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	mode, _ := errorMode(in.Meta)
	// If no inline, apply deterministic default without Lua to keep ordering stable
	if in.Meta == nil || in.Meta.Lua == nil || in.Meta.Lua.PostMapInline == "" {
		out := in
		for i, r := range in.Records {
			rec, ok := r.(Record)
			if !ok {
				if mode == "keep-going" {
					appendEnvelopeError(&out, "lua-postmap", "", "invalid record type")
					out.Records[i] = r
					continue
				}
				return Envelope{}, errors.New("lua-postmap: invalid record type")
			}
			if rec.Error != nil {
				out.Records[i] = rec
				continue
			}
			type postResult struct {
				Locator string `json:"locator"`
				Exit    int    `json:"exit,omitempty"`
			}
			pr := postResult{Locator: rec.Locator}
			if rec.Shell != nil {
				pr.Exit = rec.Shell.ExitCode
			}
			rec.Post = pr
			out.Records[i] = rec
		}
		return out, nil
	}

	code := in.Meta.Lua.PostMapInline
	if !containsReturn(code) {
		code = "return (" + code + ")"
	}

	out := in
	for i, r := range in.Records {
		rec, ok := r.(Record)
		if !ok {
			return Envelope{}, errors.New("lua-postmap: invalid record type")
		}

		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		// Minimal libs
		L.Push(L.NewFunction(lua.OpenBase))
		L.Push(lua.LString("base"))
		L.Call(1, 0)
		L.Push(L.NewFunction(lua.OpenString))
		L.Push(lua.LString("string"))
		L.Call(1, 0)
		L.Push(L.NewFunction(lua.OpenTable))
		L.Push(lua.LString("table"))
		L.Call(1, 0)
		L.Push(L.NewFunction(lua.OpenMath))
		L.Push(lua.LString("math"))
		L.Call(1, 0)

		// Globals
		L.SetGlobal("locator", lua.LString(rec.Locator))
		L.SetGlobal("meta", toLValue(L, rec.Meta))
		L.SetGlobal("mapped", toLValue(L, rec.Mapped))
		// shell as table
		var shellMap map[string]any
		if rec.Shell != nil {
			shellMap = map[string]any{
				"exitCode": rec.Shell.ExitCode,
				"stdout":   rec.Shell.Stdout,
				"stderr":   rec.Shell.Stderr,
			}
		}
		L.SetGlobal("shell", toLValue(L, shellMap))

		fn, err := L.LoadString(code)
		if err != nil {
			if mode == "keep-going" {
				appendEnvelopeError(&out, "lua-postmap", rec.Locator, err.Error())
				rec.Error = &RecError{Stage: "lua-postmap", Message: err.Error()}
				out.Records[i] = rec
				continue
			}
			return Envelope{}, fmt.Errorf("lua-postmap: %v", err)
		}
		L.Push(fn)
		done := make(chan struct{})
		var callErr error
		go func() {
			callErr = L.PCall(0, 1, nil)
			close(done)
		}()
		select {
		case <-done:
			if callErr != nil {
				if mode == "keep-going" {
					appendEnvelopeError(&out, "lua-postmap", rec.Locator, callErr.Error())
					rec.Error = &RecError{Stage: "lua-postmap", Message: callErr.Error()}
					out.Records[i] = rec
					continue
				}
				return Envelope{}, fmt.Errorf("lua-postmap: %v", callErr)
			}
		case <-time.After(200 * time.Millisecond):
			if mode == "keep-going" {
				appendEnvelopeError(&out, "lua-postmap", rec.Locator, "timeout")
				rec.Error = &RecError{Stage: "lua-postmap", Message: "timeout"}
				out.Records[i] = rec
				continue
			}
			return Envelope{}, fmt.Errorf("lua-postmap: timeout")
		}
		ret := L.Get(-1)
		L.Pop(1)
		rec.Post = fromLValue(ret)
		out.Records[i] = rec
	}
	return out, nil
}

func init() { Register("lua-postmap", luaPostMapRunner) }
