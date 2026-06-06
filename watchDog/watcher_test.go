package watchdog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	watchdog "github.com/Mritunjay2005/watchDog"
)

// helper — creates a watcher with fast debounce for tests
func newTestWatcher(t *testing.T) *watchdog.Watcher {
	t.Helper()
	w := watchdog.New(watchdog.WithDebounce(50 * time.Millisecond))
	t.Cleanup(func() { w.Stop() })
	return w
}

// Test 1 — handler fires when a file is written
func TestModifiedHandler(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t)

	got := make(chan watchdog.Event, 1)
	// on Windows, new files emit Created not Modified
	w.On(watchdog.Modified|watchdog.Created, "**/*.txt", func(e watchdog.Event) {
		got <- e
	})
	w.Start(dir)
	time.Sleep(50 * time.Millisecond)

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)

	select {
	case e := <-got:
		if !strings.HasSuffix(e.Path, "test.txt") {
			t.Errorf("unexpected path: %s", e.Path)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("handler was not called")
	}
}


// Test 2 — handler does NOT fire for non-matching pattern
func TestPatternNoMatch(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t)

	got := make(chan watchdog.Event, 1)
	w.On(watchdog.Modified, "**/*.go", func(e watchdog.Event) {
		got <- e
	})
	w.Start(dir)
	time.Sleep(50 * time.Millisecond)

	// write a .txt file — should NOT match **/*.go
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)

	select {
	case <-got:
		t.Fatal("handler should not have fired for .txt file")
	case <-time.After(300 * time.Millisecond):
		// good — no event
	}
}


// Test 3 — recursive watching fires for files in subdirectories
func TestRecursiveWatch(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subpkg")
	os.Mkdir(subdir, 0755)

	w := newTestWatcher(t)

	got := make(chan watchdog.Event, 1)
	// on Windows, new files emit Created not Modified
	w.On(watchdog.Modified|watchdog.Created, "**/*.go", func(e watchdog.Event) {
		got <- e
	})
	w.Start(dir)
	time.Sleep(50 * time.Millisecond)

	os.WriteFile(filepath.Join(subdir, "foo.go"), []byte("package subpkg"), 0644)

	select {
	case e := <-got:
		if !strings.HasSuffix(e.Path, "foo.go") {
			t.Errorf("unexpected path: %s", e.Path)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("recursive handler was not called")
	}
}
// Test 4 — debounce collapses multiple writes into one handler call
func TestDebounceCollapse(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t)

	count := make(chan struct{}, 10)
	w.On(watchdog.Modified, "**/*.txt", func(e watchdog.Event) {
		count <- struct{}{}
	})
	w.Start(dir)
	time.Sleep(50 * time.Millisecond)

	// write the same file 5 times rapidly
	path := filepath.Join(dir, "test.txt")
	for i := 0; i < 5; i++ {
		os.WriteFile(path, []byte("hello"), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// wait for debounce window to pass
	time.Sleep(300 * time.Millisecond)

	if len(count) > 2 {
		t.Errorf("expected 1-2 events after debounce, got %d", len(count))
	}
}

// Test 5 — ErrWatcherClosed returned after Stop()
func TestErrWatcherClosed(t *testing.T) {
	w := watchdog.New()
	w.Stop()

	err := w.Start(".")
	if err != watchdog.ErrWatcherClosed {
		t.Errorf("expected ErrWatcherClosed, got %v", err)
	}
}

// Test 6 — ErrInvalidPattern returned for bad glob
func TestErrInvalidPattern(t *testing.T) {
	w := watchdog.New()
	err := w.On(watchdog.Modified, "[", func(e watchdog.Event) {})
	if err != watchdog.ErrInvalidPattern {
		t.Errorf("expected ErrInvalidPattern, got %v", err)
	}
}