package main

import (
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"

	"github.com/jkratz55/konsul"
)

type AppConfig struct {
	IngressURL string `json:"ingressURL" yaml:"ingressURL"`
	Server     struct {
		Port    int `json:"port" yaml:"port"`
		Timeout int `json:"timeout" yaml:"timeout"`
	} `json:"server" yaml:"server"`
}

func main() {

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	kvClient := konsul.NewKVClient(client)
	kv, err := kvClient.Get("config/app", true)
	if err != nil {
		if errors.Is(err, konsul.ErrKeyNotFound) {
			fmt.Println("Ohhh no the configuration is missing")
		} else {
			fmt.Println("Ohhh snap something went wrong communicating with Consul")
		}
		panic(err)
	}

	cfg := AppConfig{}
	if err := kv.UnmarshalValueJSON(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("%+v", cfg)
}
