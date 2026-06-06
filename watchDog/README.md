# watchdog 🐕

> High-level file system event watching for Go

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/Mritunjay2005/watchDog.svg)](https://pkg.go.dev/github.com/Mritunjay2005/watchDog)

watchdog wraps [fsnotify](https://github.com/fsnotify/fsnotify) with everything
Go's standard tooling is missing — debouncing, glob pattern matching, recursive
watching, and typed event dispatch.

---

## Install

```bash
go get github.com/Mritunjay2005/watchDog
```

## Quickstart

```go
package main

import (
    "fmt"
    watchdog "github.com/Mritunjay2005/watchDog"
)

func main() {
    w := watchdog.New()

    w.On(watchdog.Modified|watchdog.Created, "**/*.go", func(e watchdog.Event) {
        fmt.Printf("changed: %s\n", e.Path)
    })

    w.Start(".")
    select {}
}
```

---

## Why watchdog over raw fsnotify?

| Feature | fsnotify | watchdog |
|---|---|---|
| Debouncing | ❌ | ✅ |
| Glob patterns | ❌ | ✅ |
| Recursive watching | ❌ | ✅ |
| Handler registration | ❌ | ✅ |
| Panic recovery | ❌ | ✅ |

---

## Features

- **Debouncing** — collapses multiple rapid events into one handler call
- **Glob patterns** — `**/*.go`, `config/*.yaml`, `*.md` with full `**` support
- **Recursive watching** — automatically watches all subdirectories
- **Typed dispatch** — register handlers per operation: `Created`, `Modified`, `Deleted`, `Renamed`
- **Panic recovery** — a panicking handler never crashes the watcher

---

## API Reference

### Creating a watcher

```go
w := watchdog.New(
    watchdog.WithDebounce(200 * time.Millisecond), // default: 100ms
    watchdog.WithRecursive(true),                  // default: true
    watchdog.WithIgnore("vendor", ".git"),          // default: none
)
```

### Registering handlers

```go
// single op
w.On(watchdog.Modified, "**/*.go", func(e watchdog.Event) { ... })

// combined ops
w.On(watchdog.Created|watchdog.Modified, "**/*.yaml", func(e watchdog.Event) { ... })
```

### Event fields

```go
type Event struct {
    Path string        // absolute path to the changed file
    Op   Op            // Created, Modified, Deleted, or Renamed
    Time time.Time     // when the event was detected
    Size int64         // file size in bytes
}
```

### Op constants

```go
watchdog.Created   // file was created
watchdog.Modified  // file was modified
watchdog.Deleted   // file was deleted
watchdog.Renamed   // file was renamed
```

### Errors

```go
watchdog.ErrWatcherClosed   // Start() or On() called after Stop()
watchdog.ErrInvalidPattern  // invalid glob pattern passed to On()
watchdog.ErrPathNotFound    // path passed to Start() does not exist
```

### Starting and stopping

```go
if err := w.Start("."); err != nil {
    log.Fatal(err)
}
defer w.Stop()
```

---

## Examples

| Example | Description |
|---|---|
| [basic](examples/basic/main.go) | Minimal watcher — print every .go file change |
| [hot-reload](examples/hot-reload/main.go) | Restart a subprocess when .go files change |
| [config-watch](examples/config-watch/main.go) | Auto-reload config.yaml on change |

---

## Contributing

Pull requests are welcome. For major changes please open an issue first.

## License

[MIT](LICENSE)