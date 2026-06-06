package matcher

import (
	"path"
	"strings"
)

// Match reports whether filePath matches the given glob pattern.
// Supports ** for matching across directory boundaries.
func Match(pattern, filePath string) (bool, error) {
	// normalize to forward slashes for cross-platform consistency
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	// no double star — use standard library directly
	if !strings.Contains(pattern, "**/") {
		return path.Match(pattern, filePath)
	}

	// handle ** by matching against each path suffix
	// e.g. **/*.go matches src/main.go, src/internal/foo.go, etc.
	parts := strings.SplitN(pattern, "**/", 2)
	suffix := parts[1]
	for {
		if ok, err := path.Match(suffix, filePath); ok || err != nil {
			return ok, err
		}
		idx := strings.Index(filePath, "/")
		if idx < 0 {
			break
		}
		filePath = filePath[idx+1:]
	}
	return false, nil
}