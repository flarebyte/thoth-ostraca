package stage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func stripErrorsIfNeeded(env *Envelope) {
	if env.Meta != nil && env.Meta.Errors != nil && !env.Meta.Errors.EmbedErrors {
		for i := range env.Records {
			env.Records[i].Error = nil
		}
	}
}

func stripRecordErrorIfNeeded(meta *Meta, rec *Record) {
	if rec == nil || meta == nil || meta.Errors == nil || meta.Errors.EmbedErrors {
		return
	}
	rec.Error = nil
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

type nopWriteCloser struct{ io.Writer }

func (n nopWriteCloser) Close() error { return nil }

func openWriter(outPath string) (io.WriteCloser, error) {
	if outPath == "" || outPath == "-" {
		return nopWriteCloser{Writer: os.Stdout}, nil
	}
	if dir := filepath.Dir(outPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("write-output: %v", err)
		}
	}
	f, err := os.Create(outPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func writeTo(outPath string, data []byte) error {
	w, err := openWriter(outPath)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()
	_, err = w.Write(data)
	return err
}

func writeLinesFromStream(outPath string, meta *Meta, stream <-chan Record) (bool, error) {
	w, err := openWriter(outPath)
	if err != nil {
		return false, err
	}
	defer func() { _ = w.Close() }()

	successSeen := false
	for rec := range stream {
		r := rec
		stripRecordErrorIfNeeded(meta, &r)
		if r.Error == nil {
			successSeen = true
		}
		b, encErr := encodeJSONCompact(r)
		if encErr != nil {
			return false, encErr
		}
		if _, writeErr := w.Write(b); writeErr != nil {
			return false, writeErr
		}
	}
	return successSeen, nil
}

func writeOutputRunner(_ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
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
		if deps.RecordStream != nil {
			_, err := writeLinesFromStream(outPath, env.Meta, deps.RecordStream)
			if err != nil {
				return Envelope{}, err
			}
			return in, nil
		}
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
	return in, nil
}

func init() { Register(writeOutputStage, writeOutputRunner) }
