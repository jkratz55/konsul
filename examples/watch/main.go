package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/jkratz55/konsul"
)

// AppConfig is a struct representing some imaginary configuration that might
// exist in a KV in Consul. AppConfig implements the Reloadable interface so
// it can be used with Watch function.
type AppConfig struct {
	IngressURL string `json:"ingressURL" yaml:"ingressURL"`
	Server     struct {
		Port    int `json:"port" yaml:"port"`
		Timeout int `json:"timeout" yaml:"timeout"`
	} `json:"server" yaml:"server"`
}

func (a *AppConfig) Reload(data []byte) error {
	return json.Unmarshal(data, a)
}

func main() {

	// Create Consul client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	cfg := &AppConfig{}
	err = konsul.Watch(client, "config/app", cfg, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-time.After(5 * time.Minute):
		fmt.Println("Shutting down")
	case <-ctx.Done():
		stop()
		fmt.Println("Goodbye")
	}
}
