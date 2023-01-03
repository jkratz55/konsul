package konsul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
)

// Unmarshaler is type capable of populating itself from binary data.
type Unmarshaler interface {
	Unmarshal(data []byte) error
}

// Watch watches a key in Consul's KV store and automatically refreshes a type
// with the value of the key on change.
//
// This is useful for handing configuration stored in Consul KV store and mapping
// it to a struct in Go. When the KV is changed it will automatically pass the
// value to the configuration struct implementing the Unmarshaler interface
// allowing it to refresh/update its internals.
//
// Because the nature of the Unmarshaler interface, cfg should always be a pointer.
// If a value is provided this function will have no effect as the changes will not
// be reflected.
//
// If a logger isn't provided a default one will be used.
//
// Note: By default if the call to Unmarshal on the Unmarshaler fails an error is
// logged but the object itself may be in a bad state or unusable. In some cases
// you may want your application to panic rather than continuing with a bad
// configuration state. In those cases it is recommended to panic in the Unmarshal
// method of the type being passed into Watch.
func Watch(client *api.Client, key string, cfg Unmarshaler, logger hclog.Logger) error {

	// If a logger wasn't provided use a default configured one
	if logger == nil {
		logger = hclog.Default()
	}

	plan, err := watch.Parse(map[string]any{
		"type": "key",
		"key":  key},
	)
	if err != nil {
		return err
	}

	plan.Handler = func(u uint64, raw any) {
		if raw == nil {
			return
		}
		kv, ok := raw.(*api.KVPair)
		if !ok {
			logger.Error(fmt.Sprintf("expected type *api.KVPair but got %T", raw))
			return
		}
		err := cfg.Unmarshal(kv.Value)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to unmarshall value for key %s to type %T", key, cfg),
				"error", err)
		} else {
			logger.Info(fmt.Sprintf("successfully refreshed type %T for key %s", cfg, key))
		}
	}

	go func() {
		err := plan.RunWithClientAndHclog(client, logger)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}
