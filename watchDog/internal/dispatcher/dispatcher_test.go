package dispatcher_test

import (
	"sync"
	"testing"
	"time"

	"github.com/Mritunjay2005/watchDog/internal/dispatcher"
)

func TestDispatchMatchingHandler(t *testing.T) {
	d := dispatcher.Dispatcher{}
	called := make(chan string, 1)

	d.Register(dispatcher.Handler{
		Op:      2, // Modified
		Pattern: "**/*.go",
		Fn:      func(e dispatcher.Event) { called <- e.Path },
	})

	d.Dispatch(dispatcher.Event{Path: "src/main.go", Op: 2})

	select {
	case path := <-called:
		if path != "src/main.go" {
			t.Errorf("unexpected path: %s", path)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("handler was not called")
	}
}

func TestDispatchWrongOp(t *testing.T) {
	d := dispatcher.Dispatcher{}
	called := make(chan struct{}, 1)

	d.Register(dispatcher.Handler{
		Op:      2, // Modified only
		Pattern: "**/*.go",
		Fn:      func(e dispatcher.Event) { called <- struct{}{} },
	})

	d.Dispatch(dispatcher.Event{Path: "src/main.go", Op: 1}) // Created — should not match

	select {
	case <-called:
		t.Fatal("handler should not have been called")
	case <-time.After(200 * time.Millisecond):
		// good
	}
}

func TestDispatchPanicRecovery(t *testing.T) {
	d := dispatcher.Dispatcher{}
	var wg sync.WaitGroup
	wg.Add(1)

	d.Register(dispatcher.Handler{
		Op:      2,
		Pattern: "**/*.go",
		Fn: func(e dispatcher.Event) {
			defer wg.Done()
			panic("user handler panic!")
		},
	})

	// this should not crash the program
	d.Dispatch(dispatcher.Event{Path: "main.go", Op: 2})
	wg.Wait()
}