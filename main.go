package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var CronLogger = cron.PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))
var Scheduler = cron.New(cron.WithChain(cron.SkipIfStillRunning(CronLogger)))
var Config *TaskConfig

func main() {
	logFile := initialiseLogger()
	defer logFile.Close()

	defer Scheduler.Stop()

	if config, err := LoadConfig("jobs.toml"); err != nil {
		log.Fatal(err)
	} else {
		Config = config
	}

	log.Printf("Initialising Gope for %s", Config.Title)
	log.Printf("Description: %s", Config.Description)

	go panel()

	RegisterTasks(Config)

	if len(Config.Tasks) > 0 {
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

func initialiseLogger() (file *os.File) {
	file, err := os.OpenFile("gope.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	return
}

func RegisterTasks(config *TaskConfig) {
	successfulRegisters := 0
	failedRegisters := 0

	for name := range config.Tasks {
		taskName := name
		task := config.Tasks[name]

		id, err := Scheduler.AddFunc(task.Interval, func() {
			task.Execute(taskName)
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

func panel() {
	tmpl := template.Must(template.ParseFiles("ui/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, Config)
	})
	http.ListenAndServe(":8080", nil)
}
