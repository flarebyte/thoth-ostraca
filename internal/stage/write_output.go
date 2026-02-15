package stage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const writeOutputStage = "write-output"

func getOutputSettings(meta *Meta) (outPath string, pretty bool, lines bool) {
	outPath = "-"
	if meta != nil && meta.Output != nil {
		if meta.Output.Out != "" {
			outPath = meta.Output.Out
		}
		pretty = meta.Output.Pretty
		lines = meta.Output.Lines
	}
	return
}

func hasSuccess(records []Record) bool {
	for _, r := range records {
		if r.Error == nil {
			return true
		}
	}
	return false
}

func stripErrorsIfNeeded(env *Envelope) {
	if env.Meta != nil && env.Meta.Errors != nil && !env.Meta.Errors.EmbedErrors {
		for i := range env.Records {
			env.Records[i].Error = nil
		}
	}
}

func encodeJSONCompact(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeJSONPretty(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func writeTo(outPath string, data []byte) error {
	if outPath == "" || outPath == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if dir := filepath.Dir(outPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("write-output: %v", err)
		}
	}
	return os.WriteFile(outPath, data, 0o644)
}

func writeOutputRunner(_ctx context.Context, in Envelope, _deps Deps) (Envelope, error) {
	// Prepare env for serialization
	outPath, pretty, lines := getOutputSettings(in.Meta)
	env := in
	if env.Meta == nil {
		env.Meta = &Meta{}
	}
	env.Meta.ContractVersion = "1"
	SortEnvelopeErrors(&env)
	stripErrorsIfNeeded(&env)

	if lines {
		// NDJSON lines: one per record
		var all bytes.Buffer
		for _, r := range env.Records {
			b, err := encodeJSONCompact(r)
			if err != nil {
				return Envelope{}, err
			}
			all.Write(b)
		}
		if err := writeTo(outPath, all.Bytes()); err != nil {
			return Envelope{}, err
		}
		// keep-going error semantics for lines
		if env.Meta != nil && env.Meta.Errors != nil && env.Meta.Errors.Mode == "keep-going" {
			if !hasSuccess(in.Records) {
				return Envelope{}, fmt.Errorf("keep-going: no successful records")
			}
		}
		return in, nil
	}

	// Aggregate envelope: pretty/compact
	var data []byte
	var err error
	if pretty {
		data, err = encodeJSONPretty(env)
	} else {
		data, err = encodeJSONCompact(env)
	}
	if err != nil {
		return Envelope{}, err
	}
	if err := writeTo(outPath, data); err != nil {
		return Envelope{}, err
	}
	if env.Meta != nil && env.Meta.Errors != nil && env.Meta.Errors.Mode == "keep-going" {
		if !hasSuccess(in.Records) && len(env.Errors) > 0 {
			return Envelope{}, fmt.Errorf("keep-going: no successful records")
		}
	}
	return in, nil
}

func init() { Register(writeOutputStage, writeOutputRunner) }
