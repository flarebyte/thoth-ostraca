// File Guide for dev/ai agents:
// Purpose: Turn raw Lua errors into more actionable messages by extracting line numbers and short source excerpts.
// Responsibilities:
// - Detect Lua line references in runtime or compile error text.
// - Slice a compact surrounding code excerpt for the failing line.
// - Return a sanitized multiline error message for stage error reporting.
// Architecture notes:
// - This formatter intentionally enriches only the error string; it avoids introducing a larger diagnostic object so Lua stage changes stay small.
// - Excerpts are capped to a tiny window around the failing line to keep record errors readable inside JSON output.
package stage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var luaLinePattern = regexp.MustCompile(`(?m):(\d+):`)

func formatLuaError(stage, locator, code, msg string) string {
	base := sanitizeErrorMessage(msg)
	lineNo := extractLuaLineNumber(base)
	excerpt := excerptLuaCode(code, lineNo)
	if excerpt == "" {
		return base
	}
	return fmt.Sprintf("%s\n%s", base, excerpt)
}

func extractLuaLineNumber(msg string) int {
	m := luaLinePattern.FindStringSubmatch(msg)
	if len(m) != 2 {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 1 {
		return 0
	}
	return n
}

func excerptLuaCode(code string, lineNo int) string {
	if lineNo < 1 || code == "" {
		return ""
	}
	lines := strings.Split(code, "\n")
	if lineNo > len(lines) {
		return ""
	}
	start := lineNo - 2
	if start < 0 {
		start = 0
	}
	end := lineNo + 1
	if end > len(lines) {
		end = len(lines)
	}
	out := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		prefix := " "
		if i+1 == lineNo {
			prefix = ">"
		}
		out = append(out, fmt.Sprintf("%s %d | %s", prefix, i+1, lines[i]))
	}
	return strings.Join(out, "\n")
}
