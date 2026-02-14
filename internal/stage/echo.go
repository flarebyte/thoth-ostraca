package stage

// Echo returns the same envelope, adding meta.stage = "echo".
func Echo(in Envelope) Envelope {
	if in.Meta == nil {
		in.Meta = map[string]any{}
	}
	in.Meta["stage"] = "echo"
	return in
}
