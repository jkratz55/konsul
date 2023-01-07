package main

import (
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
		panic(err)
	}

	cfg := AppConfig{}

	// This will panic if the option is None, ie there wasn't a value. It will
	// also panic if unmarshalling the value fails
	kv.Expect("key value wasn't found").
		MustUnmarshalValueJSON(&cfg)

	// If there is a value do then try to do something with it, in this case
	// unmarshall the value from JSON to a go type.
	kv.IfSome(func(val konsul.KeyValue) {
		err := val.UnmarshalValueJSON(&cfg)
		if err != nil {
			// todo: do something useful
		}
	})
}
