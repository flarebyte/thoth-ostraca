# Risks Overview (Generated)

This document summarizes key risks and mitigations.

## Summary
- Large repository and I/O scalability [performance]
- Script and shell execution safety [security]
- Configuration and scripting complexity [usability]

## Large repository and I/O scalability [performance]

- Description: Walking large trees, reading many files, and running transforms in parallel can exhaust CPU, memory, or I/O bandwidth, causing slowdowns or timeouts.
- Mitigation: Bounded worker pool, configurable timeouts/limits, deterministic ordering for aggregated output, and streaming (--lines) for better throughput.

## Script and shell execution safety [security]

- Description: Inline scripting and optional shell execution may enable command injection, unsafe environment access, or data exfiltration if misused.
- Mitigation: Sandbox Lua (limited libs), default strict templating, prefer argv over shell parsing, disable os/io by default, validate inputs, and provide secure defaults.

## Configuration and scripting complexity [usability]

- Description: Users may struggle with config structure, schema errors, or Lua script pitfalls, leading to confusion and misconfiguration.
- Mitigation: CUE schema validation with clear errors, concise examples, helpful CLI flags, and good error messages at each stage.
