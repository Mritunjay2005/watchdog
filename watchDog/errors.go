package watchdog

import "errors"

var (
    // ErrWatcherClosed is returned when On() or Start() is called
	// after Stop() has been called.
    ErrWatcherClosed = errors.New("watchdog: watcher is closed")

    // ErrInvalidPattern is returned by On() when the glob pattern
	// is not valid syntax.
    ErrInvalidPattern = errors.New("watchdog: invalid glob pattern")

    // ErrPathNotFound is returned by Start() when a watched path
	// does not exist on disk.
    ErrPathNotFound = errors.New("watchdog: path does not exist")
)