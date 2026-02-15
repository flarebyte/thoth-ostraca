package stage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"
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
	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	type res struct {
		idx   int
		rec   any
		envE  *Error
		fatal error
	}
	workers := getWorkers(in.Meta)
	jobs := make(chan int)
	results := make(chan res)
	var wg sync.WaitGroup
	var envErrs []Error
	var mu sync.Mutex
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			rec, ok := r.(Record)
			if !ok {
				if mode == "keep-going" {
					results <- res{idx: idx, rec: r, envE: &Error{Stage: "shell-exec", Message: "invalid record type"}}
					continue
				}
				results <- res{idx: idx, fatal: errors.New("shell-exec: invalid record type")}
				continue
			}
			if rec.Error != nil {
				results <- res{idx: idx, rec: rec}
				continue
			}
			mappedJSON, _ := json.Marshal(rec.Mapped)
			rendered := make([]string, len(argsT))
			for i := range argsT {
				rendered[i] = replaceJSON(argsT[i], string(mappedJSON))
			}
			cctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
			cmd := exec.CommandContext(cctx, program, rendered...)
			var outBuf, errBuf bytes.Buffer
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			runErr := cmd.Run()
			cancel()

			if cctx.Err() == context.DeadlineExceeded {
				if mode == "keep-going" {
					rec.Error = &RecError{Stage: "shell-exec", Message: "timeout"}
					results <- res{idx: idx, rec: rec, envE: &Error{Stage: "shell-exec", Locator: rec.Locator, Message: "timeout"}}
					continue
				}
				results <- res{idx: idx, fatal: fmt.Errorf("shell-exec: timeout")}
				continue
			}
			sr := &ShellResult{Stdout: outBuf.String(), Stderr: errBuf.String()}
			if runErr != nil {
				var ee *exec.ExitError
				if errors.As(runErr, &ee) {
					sr.ExitCode = ee.ExitCode()
				} else {
					if mode == "keep-going" {
						rec.Error = &RecError{Stage: "shell-exec", Message: runErr.Error()}
						results <- res{idx: idx, rec: rec, envE: &Error{Stage: "shell-exec", Locator: rec.Locator, Message: runErr.Error()}}
						continue
					}
					results <- res{idx: idx, fatal: fmt.Errorf("shell-exec: %v", runErr)}
					continue
				}
			} else {
				sr.ExitCode = 0
			}
			rec.Shell = sr
			results <- res{idx: idx, rec: rec}
		}
	}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		for i := range in.Records {
			jobs <- i
		}
		close(jobs)
	}()
	var firstErr error
	for i := 0; i < n; i++ {
		rr := <-results
		if rr.envE != nil {
			mu.Lock()
			envErrs = append(envErrs, *rr.envE)
			mu.Unlock()
		}
		if rr.fatal != nil && firstErr == nil {
			firstErr = rr.fatal
		}
		out.Records[rr.idx] = rr.rec
	}
	wg.Wait()
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
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
