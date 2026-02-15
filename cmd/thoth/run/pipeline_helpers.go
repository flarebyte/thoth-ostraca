package run

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// runStages executes the provided list of stage names in order.
func runStages(ctx context.Context, in stage.Envelope, stages []string) (stage.Envelope, error) {
	out := in
	var err error
	for _, name := range stages {
		out, err = stage.Run(ctx, name, out, stage.Deps{})
		if err != nil {
			return stage.Envelope{}, err
		}
	}
	return out, nil
}

// isLinesMode reports whether output should be rendered line-by-line.
func isLinesMode(out stage.Envelope) bool {
	return out.Meta != nil && out.Meta.Output != nil && out.Meta.Output.Lines
}

// shouldStripErrors reports whether errors should be removed from records before rendering.
func shouldStripErrors(out stage.Envelope) bool {
	return out.Meta != nil && out.Meta.Errors != nil && !out.Meta.Errors.EmbedErrors
}

// isKeepGoing reports whether keep-going mode is enabled.
func isKeepGoing(out stage.Envelope) bool {
	return out.Meta != nil && out.Meta.Errors != nil && out.Meta.Errors.Mode == "keep-going"
}

// hasAnySuccessfulRecord returns true if any record has nil Error.
func hasAnySuccessfulRecord(records []stage.Record) bool {
	for _, rec := range records {
		if rec.Error == nil {
			return true
		}
	}
	return false
}

// stripRecordError clears the Error field if the value is a stage.Record.
func stripRecordError(v stage.Record) stage.Record {
	v.Error = nil
	return v
}

// stripEnvelopeErrors clears Error in-place for all records in the envelope.
func stripEnvelopeErrors(out *stage.Envelope) {
	for i := range out.Records {
		out.Records[i].Error = nil
	}
}

// encodeJSON returns the JSON encoding string with HTML escaping disabled.
func encodeJSON(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// writeString writes s to w.
func writeString(w io.Writer, s string) error {
	_, err := fmt.Fprint(w, s)
	return err
}

// renderLines writes each record on its own line, optionally stripping errors.
func renderLines(out stage.Envelope, w io.Writer) error {
	strip := shouldStripErrors(out)
	for _, r := range out.Records {
		rec := r
		if strip {
			rec = stripRecordError(rec)
		}
		s, err := encodeJSON(rec)
		if err != nil {
			return err
		}
		if err := writeString(w, s); err != nil {
			return err
		}
	}
	if isKeepGoing(out) {
		if !hasAnySuccessfulRecord(out.Records) {
			return fmt.Errorf("keep-going: no successful records")
		}
	}
	return nil
}

// renderEnvelope writes the full envelope, optionally stripping record errors and sorting.
func renderEnvelope(out stage.Envelope, w io.Writer) error {
	if shouldStripErrors(out) {
		stripEnvelopeErrors(&out)
	}
	stage.SortEnvelopeErrors(&out)
	s, err := encodeJSON(out)
	if err != nil {
		return err
	}
	if err := writeString(w, s); err != nil {
		return err
	}
	if isKeepGoing(out) {
		if !hasAnySuccessfulRecord(out.Records) && len(out.Errors) > 0 {
			return fmt.Errorf("keep-going: no successful records")
		}
	}
	return nil
}
