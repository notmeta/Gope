package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type JobConfig struct {
	Title       string
	Description string
	Jobs        []Job `toml:"Job"`

	fileName string
}

const configFileSuffix = ".toml"

func LoadConfig() {
	err := filepath.Walk("jobs", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, configFileSuffix) {
			log.Printf("loading config file %q\n", path)

			file, err := os.Open(path)

			if err != nil {
				return err
			}
			defer file.Close()

			cfg, err := loadConfigFile(file)
			if err != nil {
				return err
			}

			// populate

			// store a map of jobs from this file to lookup after
			jobs := make(map[string]Job, len(cfg.Jobs))
			for _, j := range cfg.Jobs {
				jobs[j.Name] = j
			}

			// iterate jobs again and assign Job pointer in each Job event to preloaded Job structs
			for _, j := range cfg.Jobs {
				for on, event := range j.On { // TODO handle timeout, unknown, default

					// link event job name
					if val, ok := jobs[event.Name]; ok {
						event.job = val
					} else {
						// TODO use logger with levels and log a warning
						log.Printf("on event %q: could not find Job named %q\n", on, event.Name)
					}

					// link retry job
					if !strings.EqualFold(event.Retry.Then, "") { // only attempt to link if retry job name is not empty
						if val, ok := jobs[event.Retry.Then]; ok {
							event.Retry.job = val
						} else {
							// TODO use logger with levels and log a warning
							log.Printf("on event %q retry: could not find Job named %q\n", on, event.Retry.Then)
						}
					}

				}
				if j.IsSchedulable() {
					log.Printf("found Job %q with interval %q", j.Name, j.Interval)
					CronJobs = append(CronJobs, j)
				}
			}

			Config = append(Config, cfg)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%d files loaded:\n", len(Config))
	for _, s := range Config {
		log.Printf("%s %q - %s\n", s.fileName, s.Title, s.Description)
	}

}

func loadConfigFile(file *os.File) (cfg *JobConfig, err error) {
	var config JobConfig

	b, _ := ioutil.ReadAll(file)

	if err := toml.Unmarshal(b, &config); err != nil {
		log.Fatal(err)
		return cfg, err
	}

	config.fileName = file.Name()

	return &config, nil
}
