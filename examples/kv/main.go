package kv

import (
	"github.com/hashicorp/consul/api"

	"github.com/jkratz55/konsul"
)

func main() {

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	kvClient := konsul.NewKVClient(client)
	kv, err := kvClient.Get("config/hello", true)
	if err != nil {
		panic(err)
	}

	// This will panic if the option is None, ie there wasn't a value
	kv.Expect("key value wasn't found").Value()

	// If there is a value do then try to do something with it, in this case
	// unmarshall the value from JSON to a go type.
	var something string
	kv.IfSome(func(val konsul.KeyValue) {
		val.DecodeJSON(&something)
	})

	kv.
}
