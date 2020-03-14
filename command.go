package main

import (
	"bytes"
	"log"
	"os/exec"
	"syscall"
	"time"
)

// TODO create command output struct and use that instead of returning the below tuple
// TODO use error to return timeout event

// https://stackoverflow.com/a/40770011
func ExecuteCommand(command string, timeout int) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	var fin = make(chan struct{}, 1)

	if timeout > 0 {
		go func() {
			t := time.NewTimer(time.Duration(timeout) * time.Second)
			select {
			case <-t.C: // wait for timeout channel
				log.Println("command timed out, killing process")
				if err := cmd.Process.Kill(); err != nil {
					log.Println(err)
				}
			case <-fin: // command has finished - exit
				t.Stop() // stop our timer to patch any leaks
				return
			}
		}()
	}

	err := cmd.Run() // .Run waits for the process to finish
	stdout = outbuf.String()
	stderr = errbuf.String()

	fin <- struct{}{} // tell the channel we've finished so the timeout routine can exit

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
