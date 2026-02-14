// CUE schema for Thoth action config files

// Top-level configuration for a Thoth run.
PipelineConfig: {
  // Versioning for breaking changes
  configVersion: string & =~"^[0-9]+$" & "1"

  // Action to perform
  action: "pipeline" | "create" | "update" | "diff" | "validate"

  // File discovery options
  discovery?: {
    root?: string | "." // repo root (default ".")
    noGitignore?: bool | false // default false (respect .gitignore)
    followSymlinks?: bool | false // default false (do not follow for safety)
  }

  // Control available {file} fields for filtering/mapping
  files?: {
    info?: bool | false // include size/mode/modTime/isDir when true
    git?: bool | false  // include Git metadata when true
  }

  // Concurrency settings
  workers?: int & >=1

  // Error handling
  errors?: {
    mode?: "keep-going" | "fail-fast" | "keep-going"
    embedErrors?: bool | true
  }

  // Lua sandbox + runtime
  lua?: {
    timeoutMs?: int & >=0 | 2000
    instructionLimit?: int & >=0 | 1000000
    memoryLimitBytes?: int & >=0 | 8*1024*1024
    libs?: {
      base?: bool | true
      table?: bool | true
      string?: bool | true
      math?: bool | true
      os?: bool | false
      io?: bool | false
      coroutine?: bool | false
      debug?: bool | false
    }
    allowOSExecute?: bool | false
    allowEnv?: bool | false
    envAllowlist?: [...string]
    deterministicRandom?: bool | true
    randomSeed?: int
  }

  // Validation strictness for meta files
  validation?: {
    allowUnknownTopLevel?: bool | false
  }

  // Locator normalization policy
  locatorPolicy?: {
    allowAbsolute?: bool | false
    allowParentRefs?: bool | false
    posixStyle?: bool | true
  }

  // Diff options
  diff?: {
    includeSnapshots?: bool | false
    output?: "patch" | "summary" | "both" | "both"
  }

  // Update options
  update?: {
    merge?: "shallow" | "deep" | "jsonpatch" | "shallow"
  }

  // Filter/map/reduce scripts
  filter?: InlineScript
  map?: InlineScript
  postMap?: InlineScript
  reduce?: InlineScript

  // Shell execution
  shell?: {
    enabled?: bool | false
    program?: "bash" | "sh" | "zsh" | "bash"
    commandTemplate?: string // exactly one of commandTemplate or argsTemplate
    argsTemplate?: [...string]
    workingDir?: string | "."
    env?: [string]: string
    timeoutMs?: int & >=0 | 60000
    failFast?: bool | true
    capture?: {
      stdout?: bool | true
      stderr?: bool | true
      maxBytes?: int & >=0 | 1048576
    }
    strictTemplating?: bool | true
    killProcessGroup?: bool | true
    termGraceMs?: int & >=0 | 2000
  }

  // Output options
  output?: {
    lines?: bool | false
    pretty?: bool | false
    out?: string | "-"
  }

  // Save options (create)
  save?: {
    enabled?: bool | false
    onExists?: "ignore" | "error" | "ignore"
    dir?: string
    hashAlgo?: "sha256" | "sha256"
    hashLen?: int & >=1 | 15
  }
}

// Inline script shape
InlineScript: {
  inline: string
}

