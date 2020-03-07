package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"os"
)

type TaskConfig struct {
	Title         string
	Description   string
	Tasks         map[string]Task
	PrintOnFinish bool
}

type JobConfig struct {
	Title       string
	Description string
	Job         []job
}

type job struct {
	Name    string
	Command string
	Log     string
	Entry   bool
	Timeout int
	On      map[string]jobEvent
}

type jobEvent struct {
	Name  string
	Code  int
	Retry jobRetry
}

type jobRetry struct {
	Tries int
	Then  string
}

const (
	Unknown = "unknown"
	Timeout = "timeout"
	Default = "default"
)

func main() {
	file, err := os.Open("jobs_new.toml")

	if err != nil {
		return
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)

	var config JobConfig
	if err := toml.Unmarshal(b, &config); err != nil {
		log.Fatal(err)
	}

	log.Println(config.Job[0].On["3"].Name)

	for _, s := range config.Job {
		log.Println(s)
		fmt.Printf("%s\n", s.Name)
	}

}

//func LoadConfig(filePath string) (conf *TaskConfig, err error) {
//	file, err := os.Open(filePath)
//
//	if err != nil {
//		return nil, err
//	}
//
//	defer file.Close()
//
//	b, err := ioutil.ReadAll(file)
//	_, err = toml.Decode(string(b), &conf)
//
//	return
//}
