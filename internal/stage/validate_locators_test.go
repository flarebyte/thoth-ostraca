package stage

import "testing"

func TestNormalizeHTTPURLLocator(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "https://Example.com", want: "https://example.com/"},
		{in: "http://example.com:80/a", want: "http://example.com/a"},
		{in: "http://EXAMPLE.com:443", want: "http://example.com:443/"},
	}

	for _, tt := range tests {
		got, err := normalizeHTTPURLLocator(tt.in)
		if err != nil {
			t.Fatalf("normalizeHTTPURLLocator(%q) error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("normalizeHTTPURLLocator(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
