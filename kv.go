package konsul

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"

	"github.com/jkratz55/gonads/option"
)

type KeyValue struct {
	base *api.KVPair
}

func (kv KeyValue) Key() string {
	return kv.base.Key
}

func (kv KeyValue) Value() string {
	return string(kv.base.Value)
}

func (kv KeyValue) RawValue() []byte {
	return kv.base.Value
}

func (kv KeyValue) CreateIndex() uint64 {
	return kv.base.CreateIndex
}

func (kv KeyValue) ModifyIndex() uint64 {
	return kv.base.ModifyIndex
}

func (kv KeyValue) LockIndex() uint64 {
	return kv.base.LockIndex
}

func (kv KeyValue) Flags() uint64 {
	return kv.base.Flags
}

func (kv KeyValue) Partition() string {
	return kv.base.Partition
}

func (kv KeyValue) Namespace() string {
	return kv.base.Namespace
}

func (kv KeyValue) Session() string {
	return kv.base.Session
}

func (kv KeyValue) DecodeJSON(v any) error {
	return json.Unmarshal(kv.base.Value, v)
}

func (kv KeyValue) DecodeYAML(v any) error {
	return yaml.Unmarshal(kv.base.Value, v)
}

type KVClient struct {
	client *api.Client
}

func NewKVClient(c *api.Client) *KVClient {
	if c == nil {
		panic("a valid Consul API client must be provided")
	}
	return &KVClient{
		client: c,
	}
}

// Get retrieves a key-value from the Consul KV store. The KeyValue is returned
// wrapped by an Option as the key may or may not exist in Consul. If an error
// occurs communicating with Consul a non-nil error value will be returned.
func (c KVClient) Get(key string, allowStale bool) (option.Option[KeyValue], error) {
	kv, _, err := c.client.KV().Get(key, &api.QueryOptions{
		AllowStale: allowStale,
	})
	// Error communicating with Consul
	if err != nil {
		return option.None[KeyValue](), err
	}
	// Key doesn't exist
	if kv == nil {
		return option.None[KeyValue](), nil
	}
	return option.Some(KeyValue{base: kv}), nil
}

// MustGet retrieves a key-value from Consul KV store. If an error occurs fetching
// the key from Consul, or the key doesn't exist this will panic.
func (c KVClient) MustGet(key string, allowStale bool) KeyValue {
	kv, _, err := c.client.KV().Get(key, &api.QueryOptions{
		AllowStale: allowStale,
	})
	if err != nil {
		panic(fmt.Errorf("error retrieving key %s from Consul: %w", key, err))
	}
	if kv == nil {
		panic(fmt.Errorf("key %s doesn't exist", key))
	}
	return KeyValue{
		base: kv,
	}
}

func (c KVClient) Put(key string, value []byte) error {
	kv := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := c.client.KV().Put(kv, nil)
	return err
}

func (c KVClient) MustPut(key string, value []byte) {
	kv := &api.KVPair{
		Key:   key,
		Value: value,
	}
	if _, err := c.client.KV().Put(kv, nil); err != nil {
		panic(fmt.Errorf("failed to put KV with key %s in Consul: %w", key, err))
	}
}

func (c KVClient) PutJSON(key string, v any) error {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshalling value to JSON: %w", err)
	}
	kv := &api.KVPair{
		Key:   key,
		Value: data,
	}
	_, err = c.client.KV().Put(kv, nil)
	return err
}

func (c KVClient) MustPutJSON(key string, v any) {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(fmt.Errorf("error marshalling value to JSON: %w", err))
	}
	kv := &api.KVPair{
		Key:   key,
		Value: data,
	}
	if _, err := c.client.KV().Put(kv, nil); err != nil {
		panic(fmt.Errorf("failed to put KV with key %s in Consul: %w", key, err))
	}
}

func (c KVClient) PutYAML(key string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshalling value to YAML: %w", err)
	}
	kv := &api.KVPair{
		Key:   key,
		Value: data,
	}
	_, err = c.client.KV().Put(kv, nil)
	return err
}

func (c KVClient) MustPutYAML(key string, v any) {
	data, err := yaml.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("error marshalling value to YAML: %w", err))
	}
	kv := &api.KVPair{
		Key:   key,
		Value: data,
	}
	if _, err := c.client.KV().Put(kv, nil); err != nil {
		panic(fmt.Errorf("failed to put KV with key %s in Consul: %w", key, err))
	}
}

func (c KVClient) Delete(key string) error {
	_, err := c.client.KV().Delete(key, nil)
	return err
}
