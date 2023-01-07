package main

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"

	"github.com/jkratz55/konsul"
	zapWrapper "github.com/jkratz55/konsul/log/zap"
)

type dummyListener struct{}

func (d dummyListener) OnChange(instances []string) {
	// todo: If you are listening for changes its likely you want to store them or something ...
	fmt.Println("hello from dummy", instances)
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
	zapWrapper := zapWrapper.Wrap(logger)
	zapWrapper = zapWrapper.Named("consul.instancer")

	instancer, err := konsul.NewInstancer(konsul.InstancerConfig{
		Client:      client,
		Service:     "db-test",
		Tag:         "",
		PassingOnly: true,
		AllowStale:  true,
		Logger:      zapWrapper,
	})
	instancer.RegisterListener(dummyListener{})

	defer instancer.Close()

	// This just runs forever so we can test out registering and deregistering
	// on consul
	for {
		time.Sleep(time.Second * 5)
		fmt.Println(instancer.Instance())
		fmt.Println(instancer.Instances())
	}
}
