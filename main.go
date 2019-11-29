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

var Scheduler = cron.New()

func main() {
	defer Scheduler.Stop()

	config, err := LoadConfig("jobs.toml")

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Initialising Gope for %s", config.Title)
	log.Printf("Description: %s", config.Description)

	RegisterTasks(config)

	if len(config.Tasks) > 0 {
		Scheduler.Start()
	} else {
		log.Println("No tasks to register, holding off starting the Scheduler")
	}

	// Wait for a CTRL-C
	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func RegisterTasks(config *TaskConfig) {
	successfulRegisters := 0
	failedRegisters := 0

	for name := range config.Tasks {
		taskName := name
		task := config.Tasks[name]

		id, err := Scheduler.AddFunc(task.Interval, func() {
			_, _, exit := task.Run()

			log.Printf("Task %s finished with code %d, finding on_exit method", taskName, exit)

			onExit := task.FindOnExit(exit)

			if onExit != nil {
				log.Printf("on_exit command: %s", task.FindOnExit(exit).Command)
			} else {
				log.Printf("No on_exit found")
			}

		})

		if err != nil {
			failedRegisters++
			log.Println(errors.New(fmt.Sprintf("failed to register task: %s", err.Error())))
		} else {
			log.Printf("Successfully registered task '%s' with interval '%s'; assigned id: %d", name, task.Interval, id)
			successfulRegisters++
		}
	}

	log.Printf("%d task(s) registered.\n", successfulRegisters)
	log.Printf("%d task(s) failed to register.\n", failedRegisters)
}
