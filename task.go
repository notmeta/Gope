package main

import (
	"log"
	"math"
	"strconv"
	"time"
)

type Task struct {
	Command  string
	Interval string
	Timeout  int
	Retry    Retry
	OnExit   map[string]OnExit // key is actually an exit code int, don't be fooled
}

type Retry struct {
	Codes    []int
	Tries    int
	Delay    int
	MaxDelay int
	Backoff  float32
}

type OnExit struct {
	Command string
	OnRetry OnRetry
}

type OnRetry struct {
	Number int
}

func (t Task) FindOnExit(code int) *OnExit {
	codeStr := strconv.Itoa(code)
	if val, ok := t.OnExit[codeStr]; ok {
		return &val
	}
	return nil
}

func (t Task) Execute(taskName string) {
	_, _, exit := ExecuteCommand(t.Command)

	if Config.PrintOnFinish {
		log.Printf("Task %s finished with code %d, executing on_exit method", taskName, exit)
	}

	t.handleRetry(exit)

	onExit := t.FindOnExit(exit)

	if onExit != nil {
		log.Printf("Running on_exit command for %s; exit code: %d", taskName, exit)
		stdout, stderr, _ := ExecuteCommand(onExit.Command)

		if len(stdout) > 0 {
			log.Printf("stdout: %s", stdout)
		}

		if len(stderr) > 0 {
			log.Printf("stderr: %s", stderr)
		}
	}
}

func (t Task) shouldRetry(exit int) bool {
	return contains(t.Retry.Codes, exit)
}

func (t Task) handleRetry(initialExit int) {
	if t.Retry.Tries <= 0 {
		return
	}

	if !t.shouldRetry(initialExit) {
		return
	}

	exit := initialExit
	timeout := float32(t.Retry.Delay)
	for i := 0; i < t.Retry.Tries; i++ {
		log.Printf("Waiting %f seconds to retry", timeout)
		time.Sleep(time.Millisecond * time.Duration(timeout*1000))
		_, _, exit = ExecuteCommand(t.Command)

		if !t.shouldRetry(exit) {
			break
		}

		timeout = float32(math.Min(float64(timeout*t.Retry.Backoff), float64(t.Retry.MaxDelay)))
	}

}
