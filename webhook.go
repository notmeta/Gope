package main

import (
	"bytes"
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type CallableWebhook struct {
	Name    string
	Webhook webhook
}

type webhook struct {
	Url         string
	Payload     string
	ContentType string `toml:"content-type"`
}

func (wh CallableWebhook) execute() {
	log.Printf("executing webhook '%s'\n", wh.Name)

	_, err := http.Post(wh.Webhook.Url, wh.Webhook.ContentType, bytes.NewBuffer([]byte(wh.Webhook.Payload)))

	if err != nil {
		fmt.Println(err)
	}
}

func decode() (wh *CallableWebhook, err error) {
	var webhook CallableWebhook

	file, err := os.Open("webhook.toml")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)

	if err := toml.Unmarshal(b, &webhook); err != nil {
		log.Fatal(err)
		return wh, err
	}

	return &webhook, nil
}
