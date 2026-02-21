package stage

import "strings"

func sanitizeErrorMessage(msg string) string {
	s := strings.Join(strings.Fields(msg), " ")
	if s == "" {
		return "error"
	}
	return s
}

func sanitizedError(e Error) Error {
	e.Message = sanitizeErrorMessage(e.Message)
	return e
}
