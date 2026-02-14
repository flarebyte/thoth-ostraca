package stage

// Record is the standard per-record shape in the envelope.
// Using a struct ensures deterministic JSON field ordering.
type Record struct {
	Locator string         `json:"locator"`
	Meta    map[string]any `json:"meta,omitempty"`
	Mapped  any            `json:"mapped,omitempty"`
}
