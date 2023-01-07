package konsul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
)

// Reloadable is type capable of reloading itself from binary encoded data.
//
// Reloadable is intended to be implemented by types that represent configuration
// values stored in Consul KV that are hot-reloaded when a KV changes. When a KV
// change is detected the Reload method is called with the latest value from Consul.
type Reloadable interface {
	Reload(data []byte) error
}

// Watch watches a key in Consul's KV store and automatically refreshes a type
// with the value of the key on change.
//
// This is useful for handing configuration stored in Consul KV store and mapping
// it to a struct in Go. When the KV is changed it will automatically pass the
// value to the configuration struct implementing the Reloadable interface
// allowing it to refresh/update its internals.
//
// Because the nature of the Reloadable interface, cfg should always be a pointer.
// If a value is provided this function will have no effect as the changes will not
// be reflected.
//
// If a logger isn't provided a default one will be used.
//
// Note: By default if the call to Reload on the Reloadable fails an error is
// logged but the object itself may be in a bad state or unusable. In some cases
// you may want your application to panic rather than continuing with a bad
// configuration state. In those cases it is recommended to panic in the Reload
// method of the type being passed into Watch.
func Watch(client *api.Client, key string, cfg Reloadable, logger hclog.Logger) error {

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
		err := cfg.Reload(kv.Value)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to unmarshall value for key %s to type %T", key, cfg),
				"error", err)
		} else {
			logger.Info(fmt.Sprintf("successfully refreshed type %T for key %s", cfg, key))
		}
	}

	go func() {
		// If an error occurs and the plan to watch KVs stops running the behavior
		// of the application may become non-deterministic. For that reason this
		// will panic rather than continuing without listening for KV changes.
		err := plan.RunWithClientAndHclog(client, logger)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}
