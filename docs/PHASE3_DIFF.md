# Phase 3 vs FLOW_DESIGN.md (feat/phase3 vs main)

This document compares the intended Phase 3 scope from `docs/designs/FLOW_DESIGN.md` and the Phase 3 intent summary against what is actually implemented on branch `feat/phase3` (diffed from `main`).

## Scope Checked
- `discover-meta-files` implementation and helpers.
- `parse-validate-yaml` implementation and helpers.
- `validate-locators` implementation.
- Config/schema wiring for `validation`, `limits`, `locatorPolicy`, discovery knobs.
- E2E and stress coverage added in `script/e2e`.

## Delivered as Intended
- Real recursive discovery of `*.thoth.yaml` with deterministic sorted locators.
- `.gitignore` respected by default, with `noGitignore` opt-out.
- `followSymlinks` support with cycle-safe traversal (canonical-dir visited set).
- Permission/read errors integrated with `keep-going` vs `fail-fast`.
- Real YAML parsing (no stubs), strict required fields (`locator`, `meta`), and top-level strictness toggle (`allowUnknownTopLevel`).
- `limits.maxYAMLBytes` enforced with deterministic behavior and mode-aware errors.
- URL locator support is policy-gated (`locatorPolicy.allowURLs`) and normalized.
- Worker-pool parallelism applied and deterministic reassembly/sorting preserved.
- Stress and determinism tests added (large ingestion sets, mixed invalid/oversized/unreadable files, worker-count parity).
- `test-race` target exists in `Makefile`.

## Positive Shifts
- Determinism hardening is stronger than the minimum intent in practice: the branch has repeated-run checks and worker-cross checks in stress suites, plus deterministic error sorting.
- Discovery implementation is more defensive than a basic walk: explicit symlink-dir handling and loop prevention through canonical path tracking.
- YAML parser behavior is explicitly pinned for duplicate-key errors via targeted unit test.

## Negative Shifts / Drift
- URL normalization keeps fragments instead of stripping them.
  - Spec intent says URL normalization should strip fragments.
  - Current `normalizeHTTPURLLocator` preserves `Fragment`.
  - Impact: negative (canonicalization mismatch; potentially distinct locators for semantically same resource).
- Config version strictness remains looser than spec.
  - Spec/design expects strict version contract (examples use `"1"` and describe unknown-version rejection).
  - Runtime validation currently only requires `configVersion` to be a string; tests/configs still commonly use `"v0"`.
  - Impact: negative (contract drift and weaker config compatibility guarantees).
- Context cancellation guarantee is fully explicit in YAML parsing, but generic parallel helper paths do not uniformly expose cancellation semantics.
  - Impact: slight negative against the strongest wording of “cancellation respected” across ingestion internals.

## Neutral/Minor Naming Drift
- Action naming still uses `create-meta` / `update-meta` / `diff-meta` in runtime stage selection, while parts of full design prose use `create` / `update` / `diff`.
- Impact: neutral (behavior aligned, naming differs).

## Overall Assessment
- Phase 3 goals are substantially met and are production-leaning for discovery + YAML ingestion.
- The largest remaining mismatches are:
  - URL fragment handling in normalization.
  - Strict `configVersion` enforcement at runtime.
- Net result: **mostly positive shift**, with a few concrete correctness/contract drifts worth fixing before claiming full spec parity.
