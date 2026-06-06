// Package watchdog provides a high-level file system event library.
// It wraps fsnotify with debouncing, glob pattern matching, recursive
// watching, and typed event dispatch.
package watchdog

import "time"
// Op represents a file system operation as a bitmask.
// Multiple operations can be combined: Created | Modified.
type Op uint32
// Op constants represent the types of file system events.
const (
    Created  Op = 1 << iota // 1
    Modified                // 2
    Deleted                 // 4
    Renamed                 // 8
)
// String returns a human-readable name for the operation.
func (op Op) String() string {
    switch op {
    case Created:
        return "created"
    case Modified:
        return "modified"
    case Deleted:
        return "deleted"
    case Renamed:
        return "renamed"
    default:
        return "unknown"
    }
}

// Event represents a single file system change delivered to a handler.
type Event struct {
    // Path is the absolute path of the changed file.
	Path string
	// Op is the type of change that occurred.
	Op Op
	// Time is when the event was detected.
	Time time.Time
	// Size is the file size in bytes at the time of the event.
	Size int64
}