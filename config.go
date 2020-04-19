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

const (
	stdOutTemplate   = "${out}"
	stdErrTemplate   = "${err}"
	exitCodeTemplate = "${exit}"
)

func LoadConfig() {
	loadWebhooks()
	loadJobConfig()
}

func loadJobConfig() {
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

			cfg, err := decodeJobFile(file)
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

func decodeJobFile(file *os.File) (cfg *JobConfig, err error) {
	var config JobConfig

	b, _ := ioutil.ReadAll(file)

	if err := toml.Unmarshal(b, &config); err != nil {
		log.Fatal(err)
		return cfg, err
	}

	config.fileName = file.Name()

	return &config, nil
}

func loadWebhooks() {
	err := filepath.Walk("webhooks", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, configFileSuffix) {
			file, err := os.Open(path)

			if err != nil {
				return err
			}
			defer file.Close()

			wh, err := decodeWebhookFile(file)
			if err != nil {
				return err
			}

			Webhooks[strings.ToLower(wh.Name)] = wh
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%d webhook(s) loaded\n", len(Webhooks))

}
