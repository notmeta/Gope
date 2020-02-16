package main

import (
	"bytes"
	"log"
	"os/exec"
	"syscall"
	"time"
)

// https://stackoverflow.com/a/40770011
func ExecuteCommand(command string, timeout int) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	var ch = make(chan struct{}, 1)

	if timeout > 0 {
		go func() {
			log.Printf("waiting %d seconds for timeout\n", timeout)
			time.Sleep(time.Duration(timeout) * time.Second)

			select {
			case _ = <-ch:
				log.Println("command already finished, exiting")
			default:
				log.Println("command not finished, terminating")
				if err := cmd.Process.Kill(); err != nil {
					log.Println(err)
				}
			}
		}()
	}

	err := cmd.Run() // .Run waits for the process to finish
	stdout = outbuf.String()
	stderr = errbuf.String()

	ch <- struct{}{}

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			log.Printf("Could not get exit code for failed program: %v", command)
			exitCode = -1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return
}
