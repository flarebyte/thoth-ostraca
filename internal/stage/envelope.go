package stage

// Error represents a minimal stage error.
type Error struct {
	Stage   string `json:"stage"`
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
	FilterInline string `json:"filterInline,omitempty"`
}
