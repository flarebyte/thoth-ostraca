package stage

// Error represents a minimal stage error.
type Error struct {
	Stage   string `json:"stage"`
	Locator string `json:"locator,omitempty"`
	Message string `json:"message"`
}

// DiscoveryMeta holds discovery options.
type DiscoveryMeta struct {
	Root        string `json:"root,omitempty"`
	NoGitignore bool   `json:"noGitignore,omitempty"`
}

// ConfigMeta holds validated config essentials.
type ConfigMeta struct {
	ConfigVersion string `json:"configVersion"`
	Action        string `json:"action"`
}

// Meta holds optional metadata with deterministic JSON field order.
type Meta struct {
	Stage      string         `json:"stage,omitempty"`
	ConfigPath string         `json:"configPath,omitempty"`
	Config     *ConfigMeta    `json:"config,omitempty"`
	Discovery  *DiscoveryMeta `json:"discovery,omitempty"`
	Lua        *LuaMeta       `json:"lua,omitempty"`
	Shell      *ShellMeta     `json:"shell,omitempty"`
	Output     *OutputMeta    `json:"output,omitempty"`
	Reduced    any            `json:"reduced,omitempty"`
	Errors     *ErrorsMeta    `json:"errors,omitempty"`
}

// Envelope is a minimal JSON-serializable contract between stages.
// Field order is stable to keep JSON deterministic in tests.
type Envelope struct {
	Records []any   `json:"records"`
	Meta    *Meta   `json:"meta,omitempty"`
	Errors  []Error `json:"errors,omitempty"`
}

// LuaMeta holds minimal Lua-related settings.
type LuaMeta struct {
	FilterInline  string `json:"filterInline,omitempty"`
	MapInline     string `json:"mapInline,omitempty"`
	PostMapInline string `json:"postMapInline,omitempty"`
	ReduceInline  string `json:"reduceInline,omitempty"`
}

// ShellMeta holds minimal shell execution settings.
type ShellMeta struct {
	Enabled      bool     `json:"enabled,omitempty"`
	Program      string   `json:"program,omitempty"`
	ArgsTemplate []string `json:"argsTemplate,omitempty"`
	TimeoutMs    int      `json:"timeoutMs,omitempty"`
}

// OutputMeta holds minimal output settings.
type OutputMeta struct {
	Lines bool `json:"lines,omitempty"`
}

// ErrorsMeta holds error handling behavior.
type ErrorsMeta struct {
	Mode        string `json:"mode,omitempty"`
	EmbedErrors bool   `json:"embedErrors,omitempty"`
}
