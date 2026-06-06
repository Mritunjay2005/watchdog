package main

import (
	"fmt"
	"os"
	"time"

	watchdog "github.com/Mritunjay2005/watchDog"
)

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("error reading config:", err)
		return
	}
	fmt.Printf("config reloaded (%d bytes):\n%s\n", len(data), string(data))
}

func main() {
	w := watchdog.New(watchdog.WithDebounce(300 * time.Millisecond))

	w.On(watchdog.Modified|watchdog.Created, "**/*.yaml", func(e watchdog.Event) {
		fmt.Printf("config changed: %s\n", e.Path)
		loadConfig(e.Path)
	})

	if err := w.Start("."); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("watching for config changes... (ctrl+c to stop)")
	select {}
}