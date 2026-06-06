package watchdog

import (
	"sync"
	"time"
    "io/fs"
    "path/filepath"
    "os"
    "path"

    "github.com/Mritunjay2005/watchDog/internal/matcher"
	"github.com/Mritunjay2005/watchDog/internal/debouncer"
	"github.com/Mritunjay2005/watchDog/internal/dispatcher"
	"github.com/fsnotify/fsnotify"
)
// HandlerFunc is the function signature for all event handlers.
// It receives an Event describing the file system change.
type HandlerFunc func(Event)

type handler struct {
	pattern string
	fn      HandlerFunc
}

// Watcher watches directories and files for changes,
// routing events to registered handlers.
// Create one with New(), register handlers with On(),
// then call Start() to begin watching.
type Watcher struct {
	cfg      config
	fsw      *fsnotify.Watcher
	handlers map[Op][]handler
	done     chan struct{}
	mu       sync.RWMutex
	stopOnce sync.Once
}
// New creates a Watcher with the given options.
// If no options are provided, sensible defaults are used:
// 100ms debounce window and recursive watching enabled.
func New(opts ...Option) *Watcher {
	cfg := config{
		debounce:  100 * time.Millisecond,
		recursive: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Watcher{
		cfg:      cfg,
		handlers: make(map[Op][]handler),
		done:     make(chan struct{}),
	}
}
// On registers fn to be called when a file matching pattern
// is changed with the given Op.
// Pattern supports glob syntax including ** for recursive matching.
// Returns ErrInvalidPattern if pattern is invalid.
// Returns ErrWatcherClosed if the watcher has been stopped.
func (w *Watcher) On(op Op, pattern string, fn HandlerFunc) error {
    if w.isClosed() {
        return ErrWatcherClosed
    }
    // validate the pattern early so users get errors at registration time
    if _, err := path.Match(pattern, ""); err != nil {
        return ErrInvalidPattern
    }
    w.mu.Lock()
    defer w.mu.Unlock()
    w.handlers[op] = append(w.handlers[op], handler{pattern: pattern, fn: fn})
    return nil
}

func mapOp(op fsnotify.Op) Op {
	var out Op
	if op.Has(fsnotify.Write) {
		out |= Modified
	}
	if op.Has(fsnotify.Create) {
		out |= Created
	}
	if op.Has(fsnotify.Remove) {
		out |= Deleted
	}
	if op.Has(fsnotify.Rename) {
		out |= Renamed
	}
	return out
}

func (w *Watcher) buildDispatcher() *dispatcher.Dispatcher {
	w.mu.RLock()
	defer w.mu.RUnlock()
	d := &dispatcher.Dispatcher{}
	for op, handlers := range w.handlers {
		for _, h := range handlers {
			h := h
			d.Register(dispatcher.Handler{
				Op:      uint32(op),
				Pattern: h.pattern,
				Fn: func(e dispatcher.Event) {
					h.fn(Event{
						Path: e.Path,
						Op:   Op(e.Op),
						Time: e.Time.(time.Time),
					})
				},
			})
		}
	}
	return d
}
// Start begins watching the given paths for file system changes.
// Watching runs in the background — Start returns immediately.
// Returns ErrWatcherClosed if Stop() has already been called.
// Returns ErrPathNotFound if any path does not exist.
func (w *Watcher) Start(paths ...string) error {
	if w.isClosed() {
		return ErrWatcherClosed
	}
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return ErrPathNotFound
		}
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.fsw = fsw

	for _, p := range paths {
		if err := w.addPath(p); err != nil {
			return err
		}
	}

	d := debouncer.New(w.cfg.debounce)
	disp := w.buildDispatcher()

	go func() {
		for e := range d.Events() {
			disp.Dispatch(dispatcher.Event{
				Path: e.Path,
				Op:   e.Op,
				Time: e.Time,
			})
		}
	}()

	go func() {
		defer fsw.Close()
		for {
			select {
			case e, ok := <-fsw.Events: // ← this case was missing
				if !ok {
					return
				}
				op := mapOp(e.Op)
				if op == 0 {
					continue
				}
				// skip if file was created and already deleted
				if op == Created {
					if _, err := os.Stat(e.Name); os.IsNotExist(err) {
						continue
					}
				}
				if e.Op.Has(fsnotify.Create) {
					if info, err := os.Stat(e.Name); err == nil && info.IsDir() {
						w.fsw.Add(e.Name)
					}
				}
				d.Add(debouncer.Event{
					Path: e.Name,
					Op:   uint32(op),
					Time: time.Now(),
				})
			case err, ok := <-fsw.Errors:
				if !ok {
					return
				}
				_ = err // will route to Errors() channel in Task 17
			case <-w.done:
				return
			}
		}
	}()

	return nil
}
// Stop shuts down the watcher and releases all resources.
// It is safe to call Stop multiple times.
func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		close(w.done)
	})
}
func (w *Watcher) addPath(root string) error {
    if !w.cfg.recursive {
        return w.fsw.Add(root)
    }
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil // skip unreadable directories
        }
        if !d.IsDir() {
            return nil
        }
        // skip symlinks — following them can cause infinite loops
        info, err := os.Lstat(path)
        if err != nil {
            return nil
        }
        if info.Mode()&os.ModeSymlink != 0 {
            return filepath.SkipDir
        }
        for _, pattern := range w.cfg.ignore {
            if matched, _ := matcher.Match(pattern, d.Name()); matched {
                return filepath.SkipDir
            }
        }
        return w.fsw.Add(path)
    })
}
func (w *Watcher) isClosed() bool {
    select {
    case <-w.done:
        return true
    default:
        return false
    }
}