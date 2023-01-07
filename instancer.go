package konsul

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
)

// InstanceListener is a type the listens for changes from Instancer. An InstanceListener
// can be registered with an Instancer and upon changes Instancer will invoke the
// InstanceListener OnChange method with updated instances of the configured service.
type InstanceListener interface {
	OnChange(instances []string)
}

// InstancerConfig is a type holding the configuration properties to create and
// initialize an Instancer.
type InstancerConfig struct {
	// The Consul api Client to use to communicate with Consul. This is a required
	// field. Providing a nil value will lead to a panic.
	Client *api.Client
	// The registered service in Consul to monitor and load balance. This is a
	// required field. The default zero value will lead to a panic.
	Service string
	// An optional tag to limit the instances Instancer should consider. If this
	// value is the non zero-value only instances that have this tag will be
	// considered.
	Tag string
	// Specifies if Instancer should only consider passing/healthy instances. In
	// nearly all cases this should be set to true.
	PassingOnly bool
	// Determines how Consul client interacts with Consul servers. When true any
	// Consul server can be queried. Otherwise, all queries go to the leader.
	AllowStale bool
	// A logger to log internal behavior of Instancer. If a logger is not provided
	// a default one will be used configured at INFO level.
	Logger hclog.Logger
}

func (ic *InstancerConfig) validate() {
	if ic.Client == nil {
		panic("cannot provide nil consul api.Client, illegal use of api")
	}
	if strings.TrimSpace(ic.Service) == "" {
		panic("a consul service must be specified to load balance/monitor, illegal use of api")
	}
	if ic.Logger == nil {
		ic.Logger = hclog.Default()
	}
}

// Instancer is a client-side loadbalancer implementation based on Consul services.
// Instancer yields instances of a service registered in Consul and watches for
// changes. When changes are detected Instancer updates its internal cache of
// instances and notifies any listeners.
//
// The zero-value of Instancer is not usable. Use NewInstancer method to create
// and initialize a new Instancer.
type Instancer struct {
	client  *api.Client
	mutex   sync.RWMutex
	logger  hclog.Logger
	plan    *watch.Plan
	service string

	instances []string
	listeners []InstanceListener
	counter   uint64
}

// NewInstancer initializes a new Instancer with the provided configuration. If
// the configuration is invalid (misusing the API) this will panic. If the watch
// plan cannot be parsed this will return a non-nil error. Upon creating the
// Instancer it will begin to watch Consul for changes immediately.
//
// In the event the plan stops executing due to an error a panic will occur rather
// than continuing to run in a state where instances could be out of date/invalid.
func NewInstancer(config InstancerConfig) (*Instancer, error) {
	// Validates the configuration provided is valid and panics if the api is
	// being misused
	config.validate()

	params := map[string]any{
		"type":        "service",
		"service":     config.Service,
		"passingonly": config.PassingOnly,
		"stale":       config.AllowStale,
	}
	if config.Tag != "" {
		params["tag"] = config.Tag
	}

	plan, err := watch.Parse(params)
	if err != nil {
		return nil, fmt.Errorf("error creating watch plan for service %s: %w", config.Service, err)
	}

	instancer := &Instancer{
		client:    config.Client,
		mutex:     sync.RWMutex{},
		logger:    config.Logger,
		plan:      plan,
		instances: make([]string, 0),
		listeners: make([]InstanceListener, 0),
		counter:   0,
		service:   config.Service,
	}

	plan.Handler = instancer.handler

	go func() {
		instancer.logger.Info("Instancer is starting...",
			"Service", config.Service,
			"Tag", config.Tag,
			"PassingOnly", config.PassingOnly,
			"AllowStale", config.AllowStale)
		if err := plan.RunWithClientAndHclog(instancer.client, instancer.logger); err != nil {
			// If the plan stops running unexpected behavior may occur within the
			// application that is hard to troubleshoot/debug. In this case it's
			// better to panic rather than continuing running in a potentially bad
			// state without the callers' knowledge.
			instancer.logger.Error("plan encountered an error while executing",
				"err", err,
				"service", instancer.service)
			panic(fmt.Errorf("plan stopped running due to error: %w", err))
		}
	}()

	return instancer, nil
}

// Close stops the Instancer and the underlying Consul watch plan. After Close is
// called Instancer is not usable.
func (i *Instancer) Close() {
	i.plan.Stop()
	i.instances = make([]string, 0)
	i.listeners = make([]InstanceListener, 0)
}

// RegisterListener registers an InstanceListener with an Instancer to be notified
// when there is a changes to the instances for the configured service. Upon
// registering the OnChange method of the InstanceListener will be invoked with
// the current instances of the Instancer.
//
// Note: RegisterListener doesn't prevent the same InstanceListener from being
// registered multiple times. In such cases its OnChange method will be invoked
// multiple times.
//
// This will panic if the Instancer has been closed.
func (i *Instancer) RegisterListener(l InstanceListener) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.plan.IsStopped() {
		panic("Instancer is closed/stopped")
	}
	i.listeners = append(i.listeners, l)
	i.logger.Debug(fmt.Sprintf("Registered InstanceListener of type %T", l),
		"service", i.service)

	// Upon registration the InstanceListener is notified of the current instances
	instancesCopy := make([]string, len(i.instances))
	copy(instancesCopy, i.instances)
	l.OnChange(instancesCopy)
}

// Instance return a single instance round-robin load balanced along with a boolean
// value. If there are no instances the boolean value will be false. Otherwise, it
// will be true to indicate an instance was returned.
//
// This will panic if the Instancer has been closed.
func (i *Instancer) Instance() (string, bool) {
	if i.plan.IsStopped() {
		panic("Instancer is closed/stopped")
	}
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if len(i.instances) == 0 {
		return "", false
	}
	old := atomic.AddUint64(&i.counter, 1) - 1
	idx := old % uint64(len(i.instances))
	return i.instances[idx], true
}

// Instances returns a copy of the current set of instances
//
// This will panic if the Instancer has been closed.
func (i *Instancer) Instances() []string {
	if i.plan.IsStopped() {
		panic("Instancer is closed/stopped")
	}
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	instances := make([]string, len(i.instances))
	copy(instances, i.instances)
	return instances
}

func (i *Instancer) handler(_ uint64, data any) {
	i.logger.Info("Handler invoked, refreshing instances",
		"service", i.service)
	switch d := data.(type) {
	case []*api.ServiceEntry:
		i.mutex.Lock()
		defer i.mutex.Unlock()
		instances := make([]string, len(d))
		for j, entry := range d {
			addr := entry.Node.Address
			if entry.Service.Address != "" {
				addr = entry.Service.Address
			}
			instances[j] = fmt.Sprintf("%s:%d", addr, entry.Service.Port)
		}
		i.instances = instances
		i.logger.Info("Instances refreshed",
			"service", i.service,
			"instances", instances)

		// Notify listeners if there are any
		if len(i.listeners) > 0 {
			instancesCopy := make([]string, len(i.instances))
			copy(instancesCopy, i.instances)
			i.logger.Debug("Notifying all registered listeners",
				"service", i.service)
			for _, listener := range i.listeners {
				listener.OnChange(instancesCopy)
			}
			i.logger.Debug("All registered listeners have been notified",
				"service", i.service)
		}

	default:
		i.logger.Error(fmt.Sprintf("handler receieved unexpected type, expected *[]api.ServiceEntry but got %T", data))
	}
}
