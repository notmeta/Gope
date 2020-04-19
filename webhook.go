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

func (wh CallableWebhook) execute(prevOutput *CommandOutput) {
	payload := wh.Webhook.Payload

	if prevOutput != nil {
		payload = replaceVariable(stdOutTemplate, &prevOutput.StdOut, &payload)
		payload = replaceVariable(stdErrTemplate, &prevOutput.StdErr, &payload)
		payload = replaceVariable(exitCodeTemplate, prevOutput.ExitStr(), &payload)
	}

	_, err := http.Post(wh.Webhook.Url, wh.Webhook.ContentType, bytes.NewBuffer([]byte(payload)))

	if err != nil {
		fmt.Println(err)
	}
}

func decodeWebhookFile(file *os.File) (wh *CallableWebhook, err error) {
	var webhook CallableWebhook

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
