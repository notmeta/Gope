package main

import (
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

func main() {
	logFile := initialiseLogger()
	defer logFile.Close()
	defer Scheduler.Stop()

	log.Println("launching Gope")

	log.Println("loading config")
	LoadConfig()

	//log.Printf("Initialising Gope for %s", Config.Title)
	//log.Printf("Description: %s", Config.Description)

	log.Println("launching panel routine")
	go panel()

	//RegisterTasks(Config)
	//
	//if len(Config.Tasks) > 0 {
	//	Scheduler.Start()
	//} else {
	//	log.Println("No tasks to register, holding off starting the Scheduler")
	//}

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

func RegisterJobs(config *JobConfig) {
	//successfulRegisters := 0
	//failedRegisters := 0

	for _, job := range config.Jobs {

		fmt.Println(job.Name)

	}

}

//func RegisterTasks(config *TaskConfig) {
//	successfulRegisters := 0
//	failedRegisters := 0
//
//	for name := range config.Tasks {
//		taskName := name
//		task := config.Tasks[name]
//		task.LastExitCode = -1
//
//		config.Tasks[name] = task
//
//		id, err := Scheduler.AddFunc(task.Interval, func() {
//			exit := task.Execute(taskName)
//
//			task.LastExitCode = exit
//
//			newTime := time.Now()
//			task.LastRunTime = &newTime
//
//			config.Tasks[taskName] = task
//		})
//
//		if err != nil {
//			failedRegisters++
//			log.Println(errors.New(fmt.Sprintf("failed to register task: %s", err.Error())))
//		} else {
//			log.Printf("Successfully registered task '%s' with interval '%s'; assigned id: %d", name, task.Interval, id)
//			successfulRegisters++
//		}
//	}
//
//	log.Printf("%d task(s) registered.\n", successfulRegisters)
//	log.Printf("%d task(s) failed to register.\n", failedRegisters)
//}

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
