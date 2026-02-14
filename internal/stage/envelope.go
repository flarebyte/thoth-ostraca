package stage

// Error represents a minimal stage error.
type Error struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

// Envelope is a minimal JSON-serializable contract between stages.
type Envelope struct {
	Records []any          `json:"records"`
	Meta    map[string]any `json:"meta,omitempty"`
	Errors  []Error        `json:"errors,omitempty"`
}
