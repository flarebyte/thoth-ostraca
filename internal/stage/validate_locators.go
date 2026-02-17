package stage

import (
	"context"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
)

const validateLocatorsStage = "validate-locators"

type locatorPolicy struct {
	allowAbs   bool
	allowParen bool
	posix      bool
	allowURLs  bool
}

func policyFromMeta(meta *Meta) locatorPolicy {
	// Defaults: allowAbsolute=false, allowParentRefs=false, posixStyle=true
	p := locatorPolicy{allowAbs: false, allowParen: false, posix: true}
	if meta != nil && meta.LocatorPolicy != nil {
		p.allowAbs = meta.LocatorPolicy.AllowAbsolute
		p.allowParen = meta.LocatorPolicy.AllowParentRefs
		p.allowURLs = meta.LocatorPolicy.AllowURLs
		if meta.LocatorPolicy.PosixStyle {
			p.posix = true
		} else {
			p.posix = false
		}
	}
	return p
}

func violatesPathPolicy(loc string, p locatorPolicy) (bool, string) {
	if !p.allowAbs {
		if strings.HasPrefix(loc, "/") {
			return true, "absolute paths are not allowed"
		}
	}
	if !p.allowParen {
		// reject any '..' path segment
		segs := strings.Split(loc, "/")
		for _, s := range segs {
			if s == ".." {
				return true, "parent references ('..') are not allowed"
			}
		}
	}
	if p.posix {
		if strings.Contains(loc, "\\") {
			return true, "backslashes are not allowed in POSIX style"
		}
	}
	return false, ""
}

func parseHTTPURLLocator(loc string) (*url.URL, bool) {
	u, err := url.Parse(loc)
	if err != nil {
		return nil, false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, false
	}
	if u.Host == "" || u.Opaque != "" {
		return nil, false
	}
	return u, true
}

func normalizeHTTPURLLocator(loc string) (string, error) {
	u, ok := parseHTTPURLLocator(loc)
	if !ok {
		return "", &ErrInvalidLocator{msg: "invalid URL locator", locator: loc}
	}
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", &ErrInvalidLocator{msg: "invalid URL locator", locator: loc}
	}
	port := u.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		host = net.JoinHostPort(host, port)
	}
	out := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     u.EscapedPath(),
		RawQuery: u.RawQuery,
		Fragment: u.Fragment,
	}
	if out.Path == "" {
		out.Path = "/"
	}
	return out.String(), nil
}

type validateLocRes struct {
	idx   int
	rec   Record
	envE  *Error
	fatal error
}

func validateLocatorsRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	out := in
	mode, embed := errorMode(in.Meta)
	p := policyFromMeta(in.Meta)
	n := len(in.Records)
	jobs := make(chan int)
	results := make(chan validateLocRes)
	var wg sync.WaitGroup
	var envErrs []Error
	var mu sync.Mutex
	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			r := in.Records[idx]
			// Pass through records that already have an error
			if r.Error != nil {
				results <- validateLocRes{idx: idx, rec: r}
				continue
			}
			_, isURL := parseHTTPURLLocator(r.Locator)
			if isURL {
				if !p.allowURLs {
					msg := "URL locators are not allowed"
					if mode == "keep-going" {
						rr := r
						if embed {
							rr.Error = &RecError{Stage: validateLocatorsStage, Message: msg}
						}
						results <- validateLocRes{idx: idx, rec: rr, envE: &Error{Stage: validateLocatorsStage, Locator: r.Locator, Message: msg}}
						continue
					}
					results <- validateLocRes{idx: idx, fatal: &ErrInvalidLocator{msg: msg, locator: r.Locator}}
					continue
				}
				normalized, err := normalizeHTTPURLLocator(r.Locator)
				if err != nil {
					msg := "invalid URL locator"
					if mode == "keep-going" {
						rr := r
						if embed {
							rr.Error = &RecError{Stage: validateLocatorsStage, Message: msg}
						}
						results <- validateLocRes{idx: idx, rec: rr, envE: &Error{Stage: validateLocatorsStage, Locator: r.Locator, Message: msg}}
						continue
					}
					results <- validateLocRes{idx: idx, fatal: err}
					continue
				}
				rr := r
				rr.Locator = normalized
				results <- validateLocRes{idx: idx, rec: rr}
				continue
			}
			if bad, msg := violatesPathPolicy(r.Locator, p); bad {
				if mode == "keep-going" {
					rr := r
					if embed {
						rr.Error = &RecError{Stage: validateLocatorsStage, Message: msg}
					}
					results <- validateLocRes{idx: idx, rec: rr, envE: &Error{Stage: validateLocatorsStage, Locator: r.Locator, Message: msg}}
					continue
				}
				results <- validateLocRes{idx: idx, fatal: &ErrInvalidLocator{msg: msg, locator: r.Locator}}
				continue
			}
			results <- validateLocRes{idx: idx, rec: r}
		}
	}
	workers := getWorkers(in.Meta)
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
		if rr.rec.Locator != "" || rr.rec.Error != nil {
			out.Records[rr.idx] = rr.rec
		}
	}
	wg.Wait()
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	sort.Slice(out.Records, func(i, j int) bool {
		return out.Records[i].Locator < out.Records[j].Locator
	})
	return out, nil
}

type ErrInvalidLocator struct {
	msg     string
	locator string
}

func (e *ErrInvalidLocator) Error() string {
	return "invalid locator: " + e.msg + " (" + e.locator + ")"
}

func init() { Register("validate-locators", validateLocatorsRunner) }
