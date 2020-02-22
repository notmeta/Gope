package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net/http"
	"os"
)

type Webhook struct {
	Url     string
	Payload string
}

func sendtest() {
	wh, err := decode()

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = http.Post(wh.Url, "application/json", bytes.NewBuffer([]byte(wh.Payload)))

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func decode() (wh *Webhook, err error) {
	file, err := os.Open("webhook.toml")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)
	_, err = toml.Decode(string(b), &wh)

	return
}
