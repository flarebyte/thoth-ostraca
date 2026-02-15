# Phase 2 vs FLOW_DESIGN.md (Scope: Phase 2 only)

This document highlights where the Phase 2 implementation and guarantees differ from, or go beyond, the full design in `docs/designs/FLOW_DESIGN.md`. Focus is limited to Phase 2 features: contract hardening, new actions, centralized output, enrichment, diagnose, and determinism.

**Strict Contract Hardening**
- Output contract version: Spec documents `configVersion` but does not explicitly freeze the output envelope. Phase 2 enforces `meta.contractVersion = "1"` and validates it; contract snapshots exist. Shift: positive (stricter than spec). Action: add an “Output Contract v1” section to the spec and keep snapshots.
- Schema/tests: Spec mandates CUE validation for config; Phase 2 adds contract snapshot tests to guard against drift. Shift: positive. Action: document snapshot policy in spec.
- Error aggregation order: Spec mentions embedded errors but not ordering. Phase 2 sorts `envelope.errors` deterministically. Shift: positive. Action: specify error ordering rule in spec.

**New Actions**
- Names: Spec uses `create`, `update`, `diff`, `validate`; Phase 2 summary refers to `create-meta`, `update-meta`, `diff-meta` (fixtures adopt those names). Semantics align; naming differs. Shift: neutral. Action: standardize naming in docs (prefer short names in config/CLI; allow "-meta" in prose if helpful).
- Stage contract reuse: Both spec and Phase 2 reuse the same stage/error semantics across flows. Shift: none.

**Output Centralization**
- Dedicated writer stage: Spec materials refer to a writer (e.g., "write-report"); implementation exposes a `write-output` stage. Functionality matches (stdout/file, pretty, NDJSON). Shift: neutral naming only. Action: align naming in spec for clarity.
- NDJSON ordering: Spec allows nondeterministic line order due to parallelism; Phase 2 emits NDJSON in deterministic record order. Shift: positive. Action: consider updating spec to guarantee stable NDJSON ordering.

**Enrichment (Opt-In)**
- Flags and shape: Spec defines `files.info` and `files.git`; Phase 2 implements both with deterministic, gated outputs. Shift: none.
- Determinism: Phase 2 ensures stable field ordering and reproducible values where applicable. Spec is compatible. Shift: none.

**Diagnose Improvements**
- Prepare modes and fixture dumping: Spec describes prepare-to-boundary, `--dump-in/--dump-out`, and structured headers; Phase 2 implements these. Shift: none.

**Determinism Audit**
- Multi-run/worker counts: Spec guarantees deterministic aggregated ordering; Phase 2 adds e2e checks for byte-identical output across different worker counts and runs. Shift: positive. Action: record this stronger guarantee in spec (clarify scope for lines vs aggregated outputs).

**Summary**
- Positive shifts: output envelope versioning (`contractVersion = "1"`), deterministic error aggregation, deterministic NDJSON, and snapshot tests; stronger determinism across worker counts.
- Neutral shifts: naming differences (`create/update/diff` vs `*-meta`; `write-report` vs `write-output`).
- Follow-ups: update `FLOW_DESIGN.md` to include the output contract v1, codify error ordering, reflect deterministic NDJSON (if we keep it), and normalize naming.

