package stage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const enrichGitStage = "enrich-git"

var (
	errGitRepoNotFound       = errors.New("git repo not found")
	errGitRepoOpenFailed     = errors.New("git repo open failed")
	errGitStatusFailed       = errors.New("git status failed")
	errGitCommitLookupFailed = errors.New("git commit lookup failed")
)

func hasGitMetadataDir(root string) bool {
	p := filepath.Join(root, ".git")
	_, err := os.Stat(p)
	return err == nil
}

func repoRootFor(start string) (string, error) {
	cur, err := filepath.Abs(start)
	if err != nil {
		return "", errGitRepoOpenFailed
	}
	for {
		if hasGitMetadataDir(cur) {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", errGitRepoNotFound
		}
		cur = parent
	}
}

func enrichGitError(err error) string {
	switch {
	case errors.Is(err, errGitRepoNotFound):
		return "git repo not found"
	case errors.Is(err, errGitRepoOpenFailed):
		return "git repo open failed"
	case errors.Is(err, errGitStatusFailed):
		return "git status failed"
	case errors.Is(err, errGitCommitLookupFailed):
		return "git commit lookup failed"
	default:
		return "git error"
	}
}

func handleEnrichGitStageError(in Envelope, mode, msg string) (Envelope, error) {
	if mode == "keep-going" {
		out := in
		out.Errors = append(out.Errors, Error{Stage: enrichGitStage, Message: msg})
		SortEnvelopeErrors(&out)
		return out, nil
	}
	return Envelope{}, fmt.Errorf("%s: %s", enrichGitStage, msg)
}

func enrichGitRunner(_ context.Context, in Envelope, _ Deps) (Envelope, error) {
	if in.Meta == nil || in.Meta.Git == nil || !in.Meta.Git.Enabled {
		return in, nil
	}
	mode, embed := errorMode(in.Meta)
	root := determineRoot(in)
	repoRoot, err := repoRootFor(root)
	if err != nil {
		msg := sanitizeErrorMessage(enrichGitError(err))
		return handleEnrichGitStageError(in, mode, msg)
	}
	ctx, err := newGitContext(root, repoRoot)
	if err != nil {
		msg := sanitizeErrorMessage(enrichGitError(err))
		return handleEnrichGitStageError(in, mode, msg)
	}

	out := in
	var envErrs []Error
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		rr := r
		recGit, err := ctx.recGitFor(r.Locator)
		if err != nil {
			msg := sanitizeErrorMessage(enrichGitError(err))
			envErrs = append(envErrs, Error{Stage: enrichGitStage, Locator: r.Locator, Message: msg})
			if mode == "keep-going" {
				if embed {
					rr.Error = &RecError{Stage: enrichGitStage, Message: msg}
				}
				out.Records[i] = rr
				continue
			}
			return Envelope{}, fmt.Errorf("%s: %s", enrichGitStage, msg)
		}
		rr.Git = recGit
		out.Records[i] = rr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
		SortEnvelopeErrors(&out)
	}
	return out, nil
}

func init() { Register(enrichGitStage, enrichGitRunner) }
