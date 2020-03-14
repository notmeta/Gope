package main

type job struct {
	Name     string
	Command  string
	Interval string
	Log      string
	Entry    bool
	Timeout  int
	On       map[string]jobEvent

	//LastRunTime  *time.Time
	//LastExitCode int
}

type jobEvent struct {
	Name  string // name of job to run
	Code  int
	Retry jobRetry

	job *job
}

type jobRetry struct {
	Tries    int
	Delay    int
	MaxDelay int
	Backoff  float32
	Then     string // name of job to run after
}

const (
	Unknown = "unknown"
	Timeout = "timeout"
	Default = "default"
)
