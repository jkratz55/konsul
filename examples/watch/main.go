package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"

	"github.com/jkratz55/konsul"
	kzap "github.com/jkratz55/konsul/log/zap"
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

func (a *AppConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

func main() {

	// Create Consul client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	cfg := &AppConfig{}
	changeCount := 0
	cb := func(key string, err error) {
		changeCount++
		fmt.Println("Change Detected!")
		fmt.Printf("%+v\n", cfg)
	}

	go func() {
		err = konsul.Watch(client, "config/app", cfg, konsul.WatchOptions{
			Logger:                  kzap.Wrap(logger),
			PanicOnUnmarshalFailure: false,
			WatchNotification:       cb,
		})
		// If Watch returns an error we aren't getting KV updates anymore so we'll
		// panic rather than running in a potentially weird state because the
		// configuration changed in Consul but didn't reflect in the application.
		if err != nil {
			panic(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-ctx.Done():
		stop()
		fmt.Printf("Changes detected: %d", changeCount)
		fmt.Println("Goodbye")
	}
}
