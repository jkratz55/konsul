package konsul

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
)

// WatchNotificationFunc is a callback function that can optionally be invoked
// by Watch when a KV is changed to notify the application code. If the KV change
// was handled successfully the error value will be nil, otherwise a non-nil error
// value is passed.
type WatchNotificationFunc func(key string, err error)

// WatchOptions holds configuration properties customizing the behavior of Watch.
type WatchOptions struct {
	// The logger used to log events and errors while watching a KV in Consul.
	// If not provided a default logger will be used.
	Logger hclog.Logger
	// Flag to control if the Watch function should panic if it cannot successfully
	// unmarshall and update the target type on a KV change event. When true Watch
	// will panic the call to UnmarshalBinary returns an error.
	PanicOnUnmarshalFailure bool
	// An optional callback func that get invoked everytime a KV change is detected.
	WatchNotification WatchNotificationFunc
}

// Watch watches a key in Consul's KV store and automatically refreshes a type
// with the value of the key on change.
//
// This is useful for handing configuration stored in Consul KV store and mapping
// it to a struct in Go. When the KV is changed it will automatically pass the
// value to the configuration struct implementing the BinaryUnmarshaler interface
// allowing it to refresh/update its internals.
//
// Because the nature of the BinaryUnmarshaler interface, cfg should always be a
// pointer. If a value is provided this function will have no effect as the changes
// will not be reflected.
//
// Watch is blocking and in nearly all use cases it should be called on a new
// goroutine. Watch is intended to execute for the entire lifecycle of the
// application. It doesn't provide a mechanism to stop watching a key. It will
// only return on an error, and if it returns with an error the application will
// no longer receive updates when a KV changes. In many cases the caller may want
// to panic to prevent unexpected behavior since the configuration will not be
// updated as expected.
//
// Example:
//
//	 cfg := &AppConfig{}
//		go func() {
//			err = konsul.Watch(client, "config/app", cfg, konsul.WatchOptions{
//				Logger:                  kzap.Wrap(logger),
//				PanicOnUnmarshalFailure: false,
//			})
//			// If Watch returns an error we aren't getting KV updates anymore so we'll
//			// panic rather than running in a potentially weird state where we aren't
//			// getting updates.
//			if err != nil {
//				panic(err)
//			}
//		}()
func Watch(client *api.Client, key string, cfg encoding.BinaryUnmarshaler,
	opts WatchOptions) error {

	// If a logger is provided in the options it will be used but if one isn't
	// provided a default once is created.
	logger := hclog.Default()
	if opts.Logger != nil {
		logger = opts.Logger
	}

	// If the cfg argument isn't a pointer log out a warning as this is likely not
	// going to work as the caller intends.
	if reflect.ValueOf(cfg).Type().Kind() != reflect.Pointer {
		logger.Warn(fmt.Sprintf("cfg argument should be a pointer to a type that implements encoding.BinaryUnmarshaller interface, instead got %T. This likely will not function as the devleper intended.", cfg))
	}

	plan, err := watch.Parse(map[string]any{
		"type": "key",
		"key":  key},
	)
	if err != nil {
		return fmt.Errorf("failed to parse watch plan: %w", err)
	}

	plan.Handler = func(u uint64, raw any) {
		if raw == nil {
			return
		}
		kv, ok := raw.(*api.KVPair)
		if !ok {
			logger.Error(fmt.Sprintf("expected type *api.KVPair but got %T", raw))
			if opts.WatchNotification != nil {
				opts.WatchNotification(key, fmt.Errorf("expected type *api.KVPair but got %T", raw))
			}
			return
		}

		err := cfg.UnmarshalBinary(kv.Value)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to unmarshall value for key %s to type %T", key, cfg),
				"error", err)
			if opts.WatchNotification != nil {
				opts.WatchNotification(key, err)
			}
			if opts.PanicOnUnmarshalFailure {
				panic(err)
			}
		} else {
			logger.Info(fmt.Sprintf("successfully refreshed type %T for key %s", cfg, key))
			if opts.WatchNotification != nil {
				opts.WatchNotification(key, nil)
			}
		}
	}

	return plan.RunWithClientAndHclog(client, logger)
}
