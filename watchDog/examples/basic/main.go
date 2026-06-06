package main

import (
	"fmt"
	"time"

	watchdog "github.com/Mritunjay2005/watchDog"
)

func main() {
	w := watchdog.New(watchdog.WithDebounce(100 * time.Millisecond))

	w.On(watchdog.Modified|watchdog.Created, "**/*.go", func(e watchdog.Event) {
		fmt.Printf("changed: %s\n", e.Path)
	})

	if err := w.Start("."); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("watching for .go file changes... (ctrl+c to stop)")
	select {}
}