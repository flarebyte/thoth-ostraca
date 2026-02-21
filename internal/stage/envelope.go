package stage

import "fmt"

// Error represents a minimal stage error.
type Error struct {
	Stage   string `json:"stage"`
	Locator string `json:"locator,omitempty"`
	Message string `json:"message"`
}

// DiscoveryMeta holds discovery options.
type DiscoveryMeta struct {
	Root           string `json:"root,omitempty"`
	NoGitignore    bool   `json:"noGitignore,omitempty"`
	FollowSymlinks bool   `json:"followSymlinks,omitempty"`
}

// ConfigMeta holds validated config essentials.
type ConfigMeta struct {
	ConfigVersion string `json:"configVersion"`
	Action        string `json:"action"`
}

// Meta holds optional metadata with deterministic JSON field order.
type Meta struct {
	ContractVersion string          `json:"contractVersion,omitempty"`
	Stage           string          `json:"stage,omitempty"`
	ConfigPath      string          `json:"configPath,omitempty"`
	Config          *ConfigMeta     `json:"config,omitempty"`
	Discovery       *DiscoveryMeta  `json:"discovery,omitempty"`
	Validation      *ValidationMeta `json:"validation,omitempty"`
	Limits          *LimitsMeta     `json:"limits,omitempty"`
	LocatorPolicy   *LocatorPolicy  `json:"locatorPolicy,omitempty"`
	FileInfo        *FileInfoMeta   `json:"fileInfo,omitempty"`
	Git             *GitMeta        `json:"git,omitempty"`
	Inputs          []string        `json:"inputs,omitempty"`
	MetaFiles       []string        `json:"metaFiles,omitempty"`
	Diff            *DiffReport     `json:"diff,omitempty"`
	Lua             *LuaMeta        `json:"lua,omitempty"`
	LuaSandbox      *LuaSandboxMeta `json:"luaSandbox,omitempty"`
	Shell           *ShellMeta      `json:"shell,omitempty"`
	Output          *OutputMeta     `json:"output,omitempty"`
	UpdateMeta      *UpdateMetaMeta `json:"updateMeta,omitempty"`
	DiffMeta        *DiffMetaMeta   `json:"diffMeta,omitempty"`
	Reduced         any             `json:"reduced,omitempty"`
	Errors          *ErrorsMeta     `json:"errors,omitempty"`
	Workers         int             `json:"workers,omitempty"`
	UI              *UIMeta         `json:"ui,omitempty"`
}

// ValidationMeta controls strictness for top-level YAML fields.
type ValidationMeta struct {
	AllowUnknownTopLevel bool `json:"allowUnknownTopLevel"`
}

// LimitsMeta controls parsing size limits.
type LimitsMeta struct {
	MaxYAMLBytes       int `json:"maxYAMLBytes,omitempty"`
	MaxRecordsInMemory int `json:"maxRecordsInMemory"`
}

// DiffReport holds a minimal diff summary for meta files.
type DiffReport struct {
	OrphanMetaFiles []string     `json:"orphanMetaFiles"`
	PairedCount     int          `json:"pairedCount"`
	OrphanCount     int          `json:"orphanCount"`
	ChangedCount    int          `json:"changedCount"`
	Details         []DiffDetail `json:"details,omitempty"`
	Orphans         []string     `json:"orphans,omitempty"`
	PresentCount    int          `json:"presentCount,omitempty"`
}

// DiffDetail holds a per-locator content diff summary.
type DiffDetail struct {
	Locator         string       `json:"locator"`
	MetaFile        string       `json:"metaFile"`
	AddedKeys       []string     `json:"addedKeys"`
	RemovedKeys     []string     `json:"removedKeys"`
	ChangedKeys     []string     `json:"changedKeys"`
	TypeChangedKeys []string     `json:"typeChangedKeys,omitempty"`
	Arrays          []ArrayDiff  `json:"arrays,omitempty"`
	Changes         []DiffChange `json:"changes,omitempty"`
	Patch           []DiffOp     `json:"patch,omitempty"`
}

// ArrayDiff holds index-based array differences at a specific path.
type ArrayDiff struct {
	Path           string `json:"path"`
	AddedIndices   []int  `json:"addedIndices,omitempty"`
	RemovedIndices []int  `json:"removedIndices,omitempty"`
	ChangedIndices []int  `json:"changedIndices,omitempty"`
}

// DiffChange holds a deterministic detailed change item.
type DiffChange struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	OldValue any    `json:"oldValue,omitempty"`
	NewValue any    `json:"newValue,omitempty"`
}

// DiffOp holds one RFC 6902 JSON Patch operation.
type DiffOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// LocatorPolicy mirrors policy flags for locator validation in meta.
type LocatorPolicy struct {
	AllowAbsolute   bool `json:"allowAbsolute"`
	AllowParentRefs bool `json:"allowParentRefs"`
	PosixStyle      bool `json:"posixStyle"`
	AllowURLs       bool `json:"allowURLs"`
}

// Envelope is a minimal JSON-serializable contract between stages.
// Field order is stable to keep JSON deterministic in tests.
type Envelope struct {
	Records []Record `json:"records"`
	Meta    *Meta    `json:"meta,omitempty"`
	Errors  []Error  `json:"errors,omitempty"`
}

// ValidateEnvelope performs basic schema checks for the public contract.
// It asserts invariants used by tests to detect drift without adding new features.
func ValidateEnvelope(env Envelope) error {
	// meta.contractVersion must be present and equal to "1" in final outputs
	if env.Meta == nil || env.Meta.ContractVersion != "1" {
		return ErrInvalidContract{"missing or invalid meta.contractVersion"}
	}
	// Records slice must be non-nil (empty allowed)
	if env.Records == nil {
		return ErrInvalidContract{"records must not be null"}
	}
	// Each record must have a non-empty locator when present
	for i, r := range env.Records {
		if r.Locator == "" && r.Error == nil && r.Meta == nil && r.Mapped == nil && r.Shell == nil && r.Post == nil {
			// Allow totally empty records only if an error is embedded for keep-going
			return ErrInvalidContract{fmt.Sprintf("record[%d] missing locator", i)}
		}
	}
	return nil
}

// ErrInvalidContract is returned when envelope invariants are violated.
type ErrInvalidContract struct{ msg string }

func (e ErrInvalidContract) Error() string { return e.msg }

// LuaMeta holds minimal Lua-related settings.
type LuaMeta struct {
	FilterInline  string `json:"filterInline,omitempty"`
	MapInline     string `json:"mapInline,omitempty"`
	PostMapInline string `json:"postMapInline,omitempty"`
	ReduceInline  string `json:"reduceInline,omitempty"`
}

// LuaSandboxMeta holds runtime sandbox controls for Lua.
type LuaSandboxMeta struct {
	TimeoutMs           int                `json:"timeoutMs"`
	InstructionLimit    int                `json:"instructionLimit"`
	MemoryLimitBytes    int                `json:"memoryLimitBytes"`
	Libs                LuaSandboxLibsMeta `json:"libs"`
	DeterministicRandom bool               `json:"deterministicRandom"`
}

// LuaSandboxLibsMeta toggles exposed Lua libs.
type LuaSandboxLibsMeta struct {
	Base   bool `json:"base"`
	Table  bool `json:"table"`
	String bool `json:"string"`
	Math   bool `json:"math"`
}

// ShellMeta holds minimal shell execution settings.
type ShellMeta struct {
	Enabled          bool              `json:"enabled"`
	Program          string            `json:"program"`
	ArgsTemplate     []string          `json:"argsTemplate,omitempty"`
	WorkingDir       string            `json:"workingDir"`
	Env              map[string]string `json:"env,omitempty"`
	TimeoutMs        int               `json:"timeoutMs"`
	Capture          ShellCaptureMeta  `json:"capture"`
	StrictTemplating bool              `json:"strictTemplating"`
	KillProcessGroup bool              `json:"killProcessGroup"`
	TermGraceMs      int               `json:"termGraceMs"`
}

// ShellCaptureMeta controls shell output capture behavior.
type ShellCaptureMeta struct {
	Stdout   bool `json:"stdout"`
	Stderr   bool `json:"stderr"`
	MaxBytes int  `json:"maxBytes"`
}

// OutputMeta holds minimal output settings.
type OutputMeta struct {
	Out    string `json:"out,omitempty"`
	Pretty bool   `json:"pretty,omitempty"`
	Lines  bool   `json:"lines,omitempty"`
}

// UpdateMetaMeta holds update-meta patch settings.
type UpdateMetaMeta struct {
	Patch map[string]any `json:"patch,omitempty"`
}

// DiffMetaMeta holds diff-meta expected patch settings.
type DiffMetaMeta struct {
	ExpectedPatch map[string]any `json:"expectedPatch"`
	Format        string         `json:"format,omitempty"`
}

// ErrorsMeta holds error handling behavior.
type ErrorsMeta struct {
	Mode        string `json:"mode,omitempty"`
	EmbedErrors bool   `json:"embedErrors,omitempty"`
}

// FileInfoMeta controls file info enrichment behavior.
type FileInfoMeta struct {
	Enabled bool `json:"enabled"`
}

// GitMeta controls git enrichment behavior.
type GitMeta struct {
	Enabled bool `json:"enabled"`
}

// UIMeta holds optional runtime UI settings.
type UIMeta struct {
	Progress           bool `json:"progress"`
	ProgressIntervalMs int  `json:"progressIntervalMs"`
}
