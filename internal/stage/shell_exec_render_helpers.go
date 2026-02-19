package stage

import (
	"encoding/json"
	"fmt"
)

// renderArgs applies the {json} replacement using the record's mapped value.
func renderArgs(argsT []string, mapped any, strict bool) ([]string, error) {
	mappedJSON, _ := json.Marshal(mapped)
	rendered := make([]string, len(argsT))
	for i := range argsT {
		a := argsT[i]
		if strict {
			matches := placeholderPattern.FindAllString(a, -1)
			for _, m := range matches {
				if m != "{json}" {
					return nil, fmt.Errorf("strict templating: invalid placeholder %s", m)
				}
			}
		}
		rendered[i] = replaceJSON(a, string(mappedJSON))
	}
	return rendered, nil
}

// replaceJSON replaces exact token {json} with provided JSON string.
func replaceJSON(s, json string) string {
	out := []byte{}
	i := 0
	for i < len(s) {
		if i+6 <= len(s) && s[i:i+6] == "{json}" {
			out = append(out, []byte(json)...)
			i += 6
		} else {
			out = append(out, s[i])
			i++
		}
	}
	return string(out)
}
