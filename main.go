package main

import (
	"log"
)

func main() {

	config, err := LoadConfig("test.toml")

	if err != nil {
		log.Fatal(err)
	}

	for name, task := range config.Tasks {
		log.Printf("Running task '%s'\n", name)
		task.Run()
	}

}
