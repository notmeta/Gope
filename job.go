package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

type Job struct {
	Name     string
	Command  string
	Interval string
	Log      string
	Timeout  int
	On       map[string]*jobEvent

	LastRunTime  *time.Time
	LastExitCode int
}

type jobEvent struct {
	Name  string // name of Job to run
	Code  int
	Retry jobRetry

	job Job
}

type jobRetry struct {
	Tries    int
	Delay    int
	MaxDelay int
	Backoff  float32
	Then     string // name of Job to run after

	job Job
}

const (
	Unknown = "unknown"
	Timeout = "timeout"
	Default = "default"
)

func (j Job) IsSchedulable() bool {
	return !strings.EqualFold(j.Interval, "")
}

func (j Job) Execute() {
	if !strings.EqualFold(j.Log, "") {
		log.Println(j.Log)
	}

	if !strings.EqualFold(j.Command, "") {
		output, err := ExecuteCommand(j.Command, j.Timeout)

		if &output != nil {
			log.Printf("job finished with exit code %d", output.ExitCode)
		}

		var timeoutError *TimeoutError
		if errors.As(err, &timeoutError) {
			log.Println(err)

			if val, ok := j.On[Timeout]; ok {
				log.Printf("job %q timed out, executing on.timeout", j.Name)
				val.job.Execute()
			}

		} else {
			exitCode := strconv.Itoa(int(output.ExitCode))
			if val, ok := j.On[exitCode]; ok {
				val.job.Execute()
			}
		}
	}

}
