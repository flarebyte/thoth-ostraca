package metafile

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Marshal returns canonical YAML bytes for a thoth meta file.
func Marshal(locator string, meta map[string]any) ([]byte, error) {
	top := &yaml.Node{Kind: yaml.MappingNode}
	top.Content = append(top.Content, scalarNode("locator"), scalarFrom(locator))
	top.Content = append(top.Content, scalarNode("meta"), canonicalNode(meta))

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(top); err != nil {
		_ = enc.Close()
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	out := bytes.TrimRight(buf.Bytes(), "\n")
	out = append(out, '\n')
	return out, nil
}

// Write writes canonical YAML content to path, creating parent directories.
func Write(path, locator string, meta map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := Marshal(locator, meta)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func scalarNode(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

func scalarFrom(v any) *yaml.Node {
	n := &yaml.Node{}
	_ = n.Encode(v)
	return n
}

func canonicalNode(v any) *yaml.Node {
	switch x := v.(type) {
	case nil:
		return &yaml.Node{Kind: yaml.MappingNode}
	case map[string]any:
		return canonicalMapNode(x)
	case map[any]any:
		m := map[string]any{}
		for k, vv := range x {
			if ks, ok := k.(string); ok {
				m[ks] = vv
			}
		}
		return canonicalMapNode(m)
	case []any:
		n := &yaml.Node{Kind: yaml.SequenceNode}
		for _, it := range x {
			n.Content = append(n.Content, canonicalNode(it))
		}
		return n
	default:
		return scalarFrom(x)
	}
}

func canonicalMapNode(m map[string]any) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	if len(m) == 0 {
		return n
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		n.Content = append(n.Content, scalarNode(k), canonicalNode(m[k]))
	}
	return n
}
