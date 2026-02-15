package stage

import "context"

// Echo returns the same envelope, adding meta.stage = "echo".
func Echo(in Envelope) Envelope {
	if in.Meta == nil {
		in.Meta = &Meta{}
	}
	in.Meta.Stage = "echo"
	return in
}

func echoRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	return Echo(in), nil
}

func init() {
	Register("echo", echoRunner)
}
