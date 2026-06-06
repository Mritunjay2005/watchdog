package matcher_test

import (
	"testing"

	"github.com/Mritunjay2005/watchDog/internal/matcher"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern  string
		path     string
		expected bool
	}{
		{"**/*.go", "internal/debouncer/debouncer.go", true},
		{"**/*.go", "main.go", true},
		{"*.go", "sub/foo.go", false},
		{"*.go", "main.go", true},
		{"config/*.yaml", "config/app.yaml", true},
		{"config/*.yaml", "other/app.yaml", false},
		{"**/*.go", "internal/matcher/matcher_test.go", true},
	}

	for _, tc := range cases {
		got, err := matcher.Match(tc.pattern, tc.path)
		if err != nil {
			t.Errorf("Match(%q, %q) unexpected error: %v", tc.pattern, tc.path, err)
			continue
		}
		if got != tc.expected {
			t.Errorf("Match(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.expected)
		}
	}
}