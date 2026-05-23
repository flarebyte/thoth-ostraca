package ghflarebyte

project: {
  org:  "flarebyte"
  repo: "thoth-ostraca"
}

sync: {
  mode: "push"
}

repository: {
  description:   "CLI for metadata management in repositories"
  defaultBranch: "main"
  homepage:      "https://github.com/flarebyte/thoth-ostraca"
  visibility:    "public"
  template:      false
  topics: [
    "flarebyte",
    "go",
    "cli",
    "metadata",
  ]
  labels: [
    {
      name:        "bug"
      color:       "d73a4a"
      description: "Something isn't working"
    },
    {
      name:        "enhancement"
      color:       "a2eeef"
      description: "New feature or request"
    },
    {
      name:        "documentation"
      color:       "0075ca"
      description: "Improvements or additions to documentation"
    },
    {
      name:        "duplicate"
      color:       "cfd3d7"
      description: "This issue or pull request already exists"
    },
    {
      name:        "good first issue"
      color:       "7057ff"
      description: "Good for newcomers"
    },
    {
      name:        "help wanted"
      color:       "008672"
      description: "Extra attention is needed"
    },
    {
      name:        "invalid"
      color:       "e4e669"
      description: "This doesn't seem right"
    },
    {
      name:        "question"
      color:       "d876e3"
      description: "Further information is requested"
    },
    {
      name:        "wontfix"
      color:       "ffffff"
      description: "This will not be worked on"
    },
  ]
  features: {
    issues:              true
    wiki:                false
    projects:            false
    discussions:         false
    autoMerge:           true
    mergeCommit:         false
    rebaseMerge:         false
    squashMerge:         true
    deleteBranchOnMerge: true
  }
}

build: {
  language:             "go"
  mode:                 "binary"
  outputDir:            "build"
  checksumFile:         "build/checksums.txt"
  artifactTargetSuffix: true
  targets: [
    "darwin-arm64",
    "linux-amd64",
  ]
}

go: {
  cacheDir:    "./.gocache"
  modCacheDir: "./.gomodcache"
  toolchain:   "local"
}

devOutput: {
  color:      "auto"
  style:      "per_test"
  showPassed: true
}

coverage: {
  min: 80
  enforceMin: true
}

release: {
  versionSource:    "main.project.yaml"
  tagPrefix:        "v"
  notesMode:        "generate-notes"
  includeArtifacts: true
  artifactDir:      "build"
  includeChecksums: true
}
