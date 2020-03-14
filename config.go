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
	Jobs        []job `toml:"job"`

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

			// store a map of jobs from this file to lookup after
			jobs := make(map[string]job, len(cfg.Jobs))
			for _, j := range cfg.Jobs {
				if !strings.EqualFold(j.Interval, "") {
					log.Printf("found job %q with interval %q", j.Name, j.Interval)
				}
				jobs[j.Name] = j
			}

			// iterate jobs again and assign job pointer in each job event to preloaded job structs
			for _, j := range cfg.Jobs {
				for on, event := range j.On { // TODO handle timeout, unknown, default
					if val, ok := jobs[event.Name]; ok {
						event.job = &val
					} else {
						log.Printf("on event %q: could not find job named %q\n", on, event.Name)
					}
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
