package watchdog

import "time"

type config struct {
    debounce  time.Duration
    recursive bool
    ignore    []string
}
// Option is a functional option for configuring a Watcher.
type Option func(*config)
// WithDebounce sets the debounce window — the quiet period after
// the last event before the handler is called. Default is 100ms.
func WithDebounce(d time.Duration) Option {
    return func(c *config) { c.debounce = d }
}

// WithRecursive controls whether subdirectories are watched
// automatically. Default is true.
func WithRecursive(r bool) Option {
    return func(c *config) { c.recursive = r }
}

// WithIgnore specifies glob patterns for directories to skip
// during recursive walking. Example: WithIgnore("vendor", ".git")
func WithIgnore(patterns ...string) Option {
    return func(c *config) { c.ignore = append(c.ignore, patterns...) }
}