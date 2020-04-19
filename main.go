package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var CronLogger = cron.PrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))
var Scheduler = cron.New(cron.WithChain(cron.SkipIfStillRunning(CronLogger)))
var Config []*JobConfig
var Webhooks = make(map[string]*CallableWebhook)

func main() {
	logFile := initialiseLogger()

	defer logFile.Close()
	defer Scheduler.Stop()

	log.Println("launching Gope")

	log.Println("loading config")
	LoadConfig()

	log.Println("registering jobs with scheduler")
	RegisterJobs()

	if len(Config) > 0 {
		Scheduler.Start()
	} else {
		log.Println("No tasks to register, holding off starting the Scheduler")
	}

	log.Println("launching panel routine")
	go panel()

	// Wait for a CTRL-C
	log.Printf("now running. Press CTRL-C to exit.")
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

func RegisterJobs() {
	successfulRegisters := 0
	failedRegisters := 0

	for _, cfg := range Config {
		for index := range cfg.Jobs {
			job := &cfg.Jobs[index]

			if !job.IsSchedulable() {
				continue
			}

			// default values so they cannot be set in the config file
			job.LastExitCode = -1
			job.LastRunTime = nil

			id, err := Scheduler.AddFunc(job.Interval, func() {
				job.Execute(nil)

				t := time.Now()
				job.LastRunTime = &t
			})

			if err != nil {
				failedRegisters++
				log.Println(errors.New(fmt.Sprintf("failed to register task: %s", err.Error())))
			} else {
				log.Printf("successfully registered task '%s' with interval '%s'; assigned id: %d", job.Name, job.Interval, id)
				successfulRegisters++
			}
		}

	}

	log.Printf("%d job(s) registered.\n", successfulRegisters)
	log.Printf("%d job(s) failed to register.\n", failedRegisters)

}

func panel() {
	var allFiles []string
	files, err := ioutil.ReadDir("./ui")
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".html") {
			allFiles = append(allFiles, "./ui/"+filename)
		}
	}

	templates := template.Must(template.ParseFiles(allFiles...))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "index.html", template.FuncMap{
			"time":   time.Now(),
			"config": Config,
		})
		if err != nil {
			fmt.Println(err)
		}
	})

	http.ListenAndServe(":8080", nil)
}
