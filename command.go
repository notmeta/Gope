package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
)

type CommandOutput struct {
	StdOut   string
	StdErr   string
	ExitCode byte
}

type TimeoutError struct {
	command         string
	timeoutDuration int
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("gope timeout: command %q timed out with duration of %d", e.command, e.timeoutDuration)
}

// https://stackoverflow.com/a/40770011
func ExecuteCommand(command string, timeout int) (out CommandOutput, err error) {
	var outbuf, errbuf bytes.Buffer

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	var fin = make(chan struct{}, 1)
	var timeoutChannel = make(chan error, 1)

	if timeout > 0 {
		go func() {
			t := time.NewTimer(time.Duration(timeout) * time.Second)
			select {
			case <-t.C: // wait for timeout channel
				log.Println("command timed out, killing process")
				if err := cmd.Process.Kill(); err != nil {
					log.Println(err)
				}
				timeoutChannel <- &TimeoutError{command, timeout}
			case <-fin: // command has finished - exit
				t.Stop() // stop our timer to patch any leaks
				return
			}
		}()
	}

	err = cmd.Run() // .Run waits for the process to finish

	stdout := outbuf.String()
	stderr := errbuf.String()
	exitCode := byte(0)

	fin <- struct{}{} // tell the channel we've finished so the timeout routine can exit

	select {
	case timeoutErr := <-timeoutChannel:
		return CommandOutput{stdout, stderr, exitCode}, timeoutErr
	default:
		break
	}

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = byte(ws.ExitStatus())
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			log.Printf("Could not get exit code for failed program: %v", command)
			exitCode = 0
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = byte(ws.ExitStatus())
	}

	return CommandOutput{stdout, stderr, exitCode}, err
}
