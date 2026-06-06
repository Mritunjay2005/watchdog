package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	watchdog "github.com/Mritunjay2005/watchDog"
)

var (
	mu      sync.Mutex
	current *exec.Cmd
)

func restart(binary string) {
	mu.Lock()
	defer mu.Unlock()

	if current != nil {
		current.Process.Kill()
		current.Wait()
	}

	cmd := exec.Command(binary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println("failed to start:", err)
		return
	}
	current = cmd
	fmt.Println("restarted process")
}

func main() {
	w := watchdog.New(watchdog.WithDebounce(300 * time.Millisecond))

	w.On(watchdog.Modified|watchdog.Created, "**/*.go", func(e watchdog.Event) {
		fmt.Printf("change detected: %s — restarting...\n", e.Path)
		restart("./myapp") // replace with your binary name
	})

	if err := w.Start("."); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("hot-reload active... (ctrl+c to stop)")
	select {}
}