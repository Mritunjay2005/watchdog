package dispatcher

import (
	"github.com/Mritunjay2005/watchDog/internal/matcher"
)

type Event struct {
	Path string
	Op   uint32
	Time interface{}
}

type Handler struct {
	Op      uint32
	Pattern string
	Fn      func(Event)
}

type Dispatcher struct {
	handlers []Handler
}

// Register adds a handler to the dispatcher.
func (d *Dispatcher) Register(h Handler) {
	d.handlers = append(d.handlers, h)
}

// Dispatch routes an event to all matching handlers.
// Each handler runs in its own goroutine so slow handlers
// never block the watcher loop.
func (d *Dispatcher) Dispatch(e Event) {
	for _, h := range d.handlers {
		// check Op bitmask — does this handler care about this event type?
		if e.Op&h.Op == 0 {
			continue
		}
		// check glob pattern
		ok, err := matcher.Match(h.Pattern, e.Path)
		if err != nil || !ok {
			continue
		}
		h := h // capture loop variable — critical!
		go func() {
			defer func() { recover() }() // user panics must not crash the watcher
			h.Fn(e)
		}()
	}
}