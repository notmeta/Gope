package main

import (
	"log"
	"math"
	"strconv"
	"time"
)

type Task struct {
	Command      string
	Interval     string
	Timeout      int
	Retry        Retry
	OnExit       map[string]OnExit // key is actually an exit code int, don't be fooled
	LastRunTime  *time.Time
	LastExitCode int
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
	Log     string
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

func (t Task) Execute(taskName string) (exit int) {
	_, _, exit = ExecuteCommand(t.Command, t.Timeout)

	if Config.PrintOnFinish {
		log.Printf("Task %s finished with code %d, executing on_exit method", taskName, exit)
	}

	t.handleRetry(exit)

	onExit := t.FindOnExit(exit)

	if onExit != nil {
		if onExit.Command != "" {
			log.Printf("Running on_exit command for %s; exit code: %d", taskName, exit)
			_, _, _ = ExecuteCommand(onExit.Command, 0)
		}

		if onExit.Log != "" { // if there is a log statement
			log.Printf("'%s' onexit (%d) log: %s", taskName, exit, onExit.Log)
		}
	}

	return
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
		_, _, exit = ExecuteCommand(t.Command, t.Timeout)

		if !t.shouldRetry(exit) {
			break
		}

		timeout = float32(math.Min(float64(timeout*t.Retry.Backoff), float64(t.Retry.MaxDelay)))
	}

}
