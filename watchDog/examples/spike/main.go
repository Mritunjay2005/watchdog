package main

import (
    "fmt"
    "log"
    "github.com/fsnotify/fsnotify"
)

func main() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    if err := watcher.Add("."); err != nil {
        log.Fatal(err)
    }

    fmt.Println("watching... edit any file")
    for event := range watcher.Events {
        fmt.Printf("op=%-10s path=%s\n", event.Op, event.Name)
    }
}

