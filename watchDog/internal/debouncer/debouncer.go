package debouncer

import (
    "sync"
    "time"
)

type Event struct {
    Path string
    Op   uint32
    Time time.Time
}

type Debouncer struct {
    window time.Duration
    timers map[string]*time.Timer
    out    chan Event
    mu     sync.Mutex
}

func New(window time.Duration) *Debouncer {
    return &Debouncer{
        window: window,
        timers: make(map[string]*time.Timer),
        out:    make(chan Event, 64), // buffered — never block the caller
    }
}

func (d *Debouncer) Add(e Event) {
    d.mu.Lock()
    defer d.mu.Unlock()
    if t, ok := d.timers[e.Path]; ok {
        t.Reset(d.window) // file still changing — push the timer forward
        return
    }
    // first event for this path — start a timer
    d.timers[e.Path] = time.AfterFunc(d.window, func() {
        d.mu.Lock()
        delete(d.timers, e.Path)
        d.mu.Unlock()
        d.out <- e
    })
}

func (d *Debouncer) Events() <-chan Event { return d.out }