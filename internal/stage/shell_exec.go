package stage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

func shellExecRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// If not enabled, passthrough
	enabled := false
	program := "bash"
	var argsT []string
	timeout := 60000
	if in.Meta != nil && in.Meta.Shell != nil {
		enabled = in.Meta.Shell.Enabled
		if in.Meta.Shell.Program != "" {
			program = in.Meta.Shell.Program
		}
		if len(in.Meta.Shell.ArgsTemplate) > 0 {
			argsT = in.Meta.Shell.ArgsTemplate
		}
		if in.Meta.Shell.TimeoutMs > 0 {
			timeout = in.Meta.Shell.TimeoutMs
		}
	}
	if !enabled {
		return in, nil
	}

	out := in
	for i, r := range in.Records {
		rec, ok := r.(Record)
		if !ok {
			return Envelope{}, errors.New("shell-exec: invalid record type")
		}
		// Render args
		mappedJSON, _ := json.Marshal(rec.Mapped)
		rendered := make([]string, len(argsT))
		for i := range argsT {
			rendered[i] = replaceJSON(argsT[i], string(mappedJSON))
		}

		// Execute with timeout
		cctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()
		cmd := exec.CommandContext(cctx, program, rendered...)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		runErr := cmd.Run()

		if cctx.Err() == context.DeadlineExceeded {
			return Envelope{}, fmt.Errorf("shell-exec: timeout")
		}

		sr := &ShellResult{Stdout: outBuf.String(), Stderr: errBuf.String()}
		if runErr != nil {
			var ee *exec.ExitError
			if errors.As(runErr, &ee) {
				sr.ExitCode = ee.ExitCode()
			} else {
				return Envelope{}, fmt.Errorf("shell-exec: %v", runErr)
			}
		} else {
			sr.ExitCode = 0
		}
		rec.Shell = sr
		out.Records[i] = rec
	}
	return out, nil
}

func replaceJSON(s, json string) string {
	// Replace exactly the token {json}
	// No templating beyond this.
	// Simple non-allocating approach
	out := []byte{}
	i := 0
	for i < len(s) {
		if i+6 <= len(s) && s[i:i+6] == "{json}" {
			out = append(out, []byte(json)...)
			i += 6
		} else {
			out = append(out, s[i])
			i++
		}
	}
	return string(out)
}

func init() { Register("shell-exec", shellExecRunner) }
