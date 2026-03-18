package stage

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// renderArgs applies supported placeholders using the current record context.
func renderArgs(argsT []string, rec Record, strict bool) ([]string, error) {
	rendered := make([]string, len(argsT))
	for i := range argsT {
		a := argsT[i]
		out, err := renderArg(a, rec, strict)
		if err != nil {
			return nil, err
		}
		rendered[i] = out
	}
	return rendered, nil
}

func renderArg(s string, rec Record, strict bool) (string, error) {
	out := []byte{}
	i := 0
	for i < len(s) {
		if s[i] != '{' {
			out = append(out, s[i])
			i++
			continue
		}
		end := strings.IndexByte(s[i:], '}')
		if end <= 0 {
			out = append(out, s[i])
			i++
			continue
		}
		placeholder := s[i : i+end+1]
		value, handled, err := resolvePlaceholder(placeholder, rec)
		if err != nil {
			return "", err
		}
		if !handled {
			if strict {
				return "", fmt.Errorf(
					"strict templating: invalid placeholder %s",
					placeholder,
				)
			}
			out = append(out, placeholder...)
			i += end + 1
			continue
		}
		out = append(out, value...)
		i += end + 1
	}
	return string(out), nil
}

func resolvePlaceholder(
	placeholder string,
	rec Record,
) (string, bool, error) {
	switch placeholder {
	case "{json}":
		mappedJSON, _ := json.Marshal(rec.Mapped)
		return string(mappedJSON), true, nil
	case "{locator}":
		return rec.Locator, true, nil
	case "{file.base}":
		return path.Base(rec.Locator), true, nil
	case "{file.dir}":
		return path.Dir(rec.Locator), true, nil
	case "{file.stem}":
		base := path.Base(rec.Locator)
		ext := path.Ext(base)
		return strings.TrimSuffix(base, ext), true, nil
	case "{file.ext}":
		return path.Ext(rec.Locator), true, nil
	}
	if strings.HasPrefix(placeholder, "{mapped.") &&
		strings.HasSuffix(placeholder, "}") {
		keyPath := strings.TrimSuffix(
			strings.TrimPrefix(placeholder, "{mapped."),
			"}",
		)
		v, err := lookupMappedValue(rec.Mapped, keyPath)
		if err != nil {
			return "", true, err
		}
		return placeholderValueString(v)
	}
	return "", false, nil
}

func lookupMappedValue(mapped any, keyPath string) (any, error) {
	cur := mapped
	for _, part := range strings.Split(keyPath, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"template placeholder {mapped.%s} requires object value",
				keyPath,
			)
		}
		next, ok := m[part]
		if !ok {
			return nil, fmt.Errorf(
				"template placeholder {mapped.%s} missing value",
				keyPath,
			)
		}
		cur = next
	}
	return cur, nil
}

func placeholderValueString(v any) (string, bool, error) {
	switch x := v.(type) {
	case nil:
		return "", true, fmt.Errorf("template placeholder missing value")
	case string:
		return x, true, nil
	case bool:
		if x {
			return "true", true, nil
		}
		return "false", true, nil
	case int:
		return fmt.Sprintf("%d", x), true, nil
	case int64:
		return fmt.Sprintf("%d", x), true, nil
	case float64:
		return fmt.Sprintf("%g", x), true, nil
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return "", true, err
		}
		return string(b), true, nil
	}
}
