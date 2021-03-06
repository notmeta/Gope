package main

import (
	"errors"
	"log"
	"strings"
	"time"
)

type Job struct {
	Name     string
	Command  string
	Interval string
	Log      string
	Webhook  string
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

func isBlank(input *string) bool {
	return strings.EqualFold(strings.TrimSpace(*input), "")
}

func (j Job) IsSchedulable() bool {
	return !isBlank(&j.Interval)
}

func (j Job) Execute(previousOutput *CommandOutput) {
	// log statement
	if !isBlank(&j.Log) {
		log.Println(j.Log)
	}

	// command statement
	if !isBlank(&j.Command) {
		output, err := ExecuteCommand(j.Command, j.Timeout)

		if &output != nil {
			log.Printf("job finished with exit code %d", output.ExitCode)
		}

		var timeoutError *TimeoutError
		if errors.As(err, &timeoutError) {
			log.Println(err)

			if val, ok := j.On[Timeout]; ok {
				log.Printf("job %q timed out, executing on.timeout", j.Name)
				val.job.Execute(&output)
			}
		} else {
			if val, ok := j.On[*output.ExitStr()]; ok {
				val.job.Execute(&output)
			} else {
				if val, ok := j.On[Unknown]; ok {
					log.Printf("no event found for exit code %d, executing on.unknown event\n", output.ExitCode)
					val.job.Execute(&output)
				}
				// !ok - no event found
			}
		}
	}

	// webhook statement
	if !isBlank(&j.Webhook) {
		webhookName := strings.ToLower(j.Webhook)
		if val, ok := Webhooks[webhookName]; ok {
			log.Printf("executing webhook %q\n", webhookName)
			val.execute(previousOutput)
		}
	}

}
