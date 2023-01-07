package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/jkratz55/konsul"
)

type AppConfig struct {
	IngressURL string `json:"ingressURL"`
	Server     struct {
		Port    int `json:"port"`
		Timeout int `json:"timeout"`
	} `json:"server"`
}

func (a *AppConfig) Reload(data []byte) error {
	return json.Unmarshal(data, a)
}

func main() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	cfg := &AppConfig{}

	err = konsul.Watch(client, "config/app", cfg, nil)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Minute * 1)
		fmt.Println(cfg)
	}
}
