package debouncer_test

import (
    "sync"
    "testing"
    "time"

    "github.com/Mritunjay2005/watchDog/internal/debouncer"
)

// Test 1 — burst collapse: 10 rapid events should produce exactly 1 output
func TestBurstCollapse(t *testing.T) {
    d := debouncer.New(100 * time.Millisecond)

    for i := 0; i < 10; i++ {
        d.Add(debouncer.Event{Path: "foo.go", Op: 2})
        time.Sleep(5 * time.Millisecond)
    }

    // expect exactly one event
    select {
    case <-d.Events():
        // good
    case <-time.After(500 * time.Millisecond):
        t.Fatal("expected one event, got none")
    }

    // ensure no second event arrives
    select {
    case <-d.Events():
        t.Fatal("expected only one event, got two")
    case <-time.After(300 * time.Millisecond):
        // good
    }
}

// Test 2 — two different files should produce two separate events
func TestTwoFiles(t *testing.T) {
    d := debouncer.New(100 * time.Millisecond)

    d.Add(debouncer.Event{Path: "foo.go", Op: 2})
    d.Add(debouncer.Event{Path: "bar.go", Op: 2})

    got := map[string]bool{}
    for i := 0; i < 2; i++ {
        select {
        case e := <-d.Events():
            got[e.Path] = true
        case <-time.After(500 * time.Millisecond):
            t.Fatal("timed out waiting for events")
        }
    }

    if !got["foo.go"] || !got["bar.go"] {
        t.Errorf("expected both files, got: %v", got)
    }
}

// Test 3 — two events separated by the full window should produce two outputs
func TestTimingWindow(t *testing.T) {
    d := debouncer.New(100 * time.Millisecond)

    d.Add(debouncer.Event{Path: "foo.go", Op: 2})
    time.Sleep(250 * time.Millisecond) // wait past the window
    d.Add(debouncer.Event{Path: "foo.go", Op: 2})

    for i := 0; i < 2; i++ {
        select {
        case <-d.Events():
            // good
        case <-time.After(500 * time.Millisecond):
            t.Fatalf("expected 2 events, only got %d", i)
        }
    }
}

// Test 4 — concurrent Add() calls must not panic or race
func TestConcurrency(t *testing.T) {
    d := debouncer.New(100 * time.Millisecond)

    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            d.Add(debouncer.Event{Path: "foo.go", Op: 2})
        }()
    }
    wg.Wait()
}