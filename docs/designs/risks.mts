import type { Risk } from './common.mts';

// Central risk catalogue. Extend as the design evolves.
export const risks: Record<string, Risk> = {
  security: {
    name: 'security',
    title: 'Script and shell execution safety',
    description:
      'Inline scripting and optional shell execution may enable command injection, unsafe environment access, or data exfiltration if misused.',
    mitigation:
      'Sandbox Lua (limited libs), default strict templating, prefer argv over shell parsing, disable os/io by default, validate inputs, and provide secure defaults.',
  },
  performance: {
    name: 'performance',
    title: 'Large repository and I/O scalability',
    description:
      'Walking large trees, reading many files, and running transforms in parallel can exhaust CPU, memory, or I/O bandwidth, causing slowdowns or timeouts.',
    mitigation:
      'Bounded worker pool, configurable timeouts/limits, deterministic ordering for aggregated output, and streaming (--lines) for better throughput.',
  },
  usability: {
    name: 'usability',
    title: 'Configuration and scripting complexity',
    description:
      'Users may struggle with config structure, schema errors, or Lua script pitfalls, leading to confusion and misconfiguration.',
    mitigation:
      'CUE schema validation with clear errors, concise examples, helpful CLI flags, and good error messages at each stage.',
  },
};
