package stage

// RecError is a per-record error payload.
type RecError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

func errorMode(meta *Meta) (mode string, embed bool) {
	mode = "fail-fast"
	if meta != nil && meta.Errors != nil {
		if meta.Errors.Mode != "" {
			mode = meta.Errors.Mode
		}
		embed = meta.Errors.EmbedErrors
	}
	return
}
