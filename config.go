package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type TaskConfig struct {
	Title         string
	Description   string
	Tasks         map[string]Task
	PrintOnFinish bool
}

func LoadConfig(filePath string) (conf *TaskConfig, err error) {
	file, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)
	_, err = toml.Decode(string(b), &conf)

	return
}
