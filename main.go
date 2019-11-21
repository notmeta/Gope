package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	config, err := LoadConfig("test.toml")

	if err != nil {
		log.Fatal(err)
	}

	successfulRegisters := 0
	failedRegisters := 0
	c := cron.New()

	for name := range config.Tasks {
		task := config.Tasks[name]

		id, err := c.AddFunc(task.Interval, func() {
			task.Run()
		})

		if err != nil {
			failedRegisters++
			log.Println(errors.New(fmt.Sprintf("failed to register task: %s", err.Error())))
		} else {
			log.Printf("Successfully registered task '%s' with interval '%s'; assigned id: %d", name, task.Interval, id)
			successfulRegisters++
		}
	}

	if len(config.Tasks) > 0 {
		c.Start()
	} else {
		log.Println("No tasks to register, holding off starting the scheduler")
	}

	log.Printf("%d task(s) registered.\n", successfulRegisters)
	log.Printf("%d task(s) failed to register.\n", failedRegisters)

	// Wait for a CTRL-C
	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	c.Stop()

}
