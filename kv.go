package konsul

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

var (
	// ErrKeyNotFound is a sentinel error value indicating the key provided
	// doesn't exist.
	ErrKeyNotFound = errors.New("key not found")
)

// KeyValue is a wrapper around KVPair type from official Consul API package.
// It provides convenient methods to unmarshal the value from Consul as JSON
// or YAML to a Go type.
type KeyValue struct {
	base *api.KVPair
}

// Key is the name of the key. It is also part of the URL path when accessed
// via the API.
func (kv KeyValue) Key() string {
	return kv.base.Key
}

// Value is the value for the key represented as a string.
func (kv KeyValue) Value() string {
	return string(kv.base.Value)
}

// RawValue is the value for the key. This can be any value and is represented
// as bytes.
func (kv KeyValue) RawValue() []byte {
	return kv.base.Value
}

// CreateIndex holds the index corresponding the creation of this KVPair. This
// is a read-only field.
func (kv KeyValue) CreateIndex() uint64 {
	return kv.base.CreateIndex
}

// ModifyIndex is used for the Check-And-Set operations and can also be fed back
// into the WaitIndex of the QueryOptions in order to perform blocking queries.
func (kv KeyValue) ModifyIndex() uint64 {
	return kv.base.ModifyIndex
}

// LockIndex holds the index corresponding to a lock on this key, if any. This is
// a read-only field.
func (kv KeyValue) LockIndex() uint64 {
	return kv.base.LockIndex
}

// Flags are any user-defined flags on the key. It is up to the implementer to check
// these values, since Consul does not treat them specially.
func (kv KeyValue) Flags() uint64 {
	return kv.base.Flags
}

// Partition is the partition the KVPair is associated with Admin Partition is a
// Consul Enterprise feature.
func (kv KeyValue) Partition() string {
	return kv.base.Partition
}

// Namespace is the namespace the KVPair is associated with Namespacing is a Consul
// Enterprise feature.
func (kv KeyValue) Namespace() string {
	return kv.base.Namespace
}

// Session is a string representing the ID of the session. Any other interactions
// with this key over the same session must specify the same session ID.
func (kv KeyValue) Session() string {
	return kv.base.Session
}

// IsEmpty returns a bool indicating if the value of the KV is empty.
//
// IsEmpty can be helpful for handling cases where the key exists in Consul KV
// store but could have an empty value.
func (kv KeyValue) IsEmpty() bool {
	return len(kv.base.Value) == 0
}

// UnmarshalValueJSON parses the JSON-encoded data of the KeyValue and stores the
// result in the value pointed to by v. If v is nil or not a pointer, UnmarshalValueJSON
// returns an InvalidUnmarshalError.
func (kv KeyValue) UnmarshalValueJSON(v any) error {
	return json.Unmarshal(kv.base.Value, v)
}

// MustUnmarshalValueJSON parses the JSON-encoded data of the KeyValue and stores the
// result in the value pointed to by v. If an error occurs during unmarshalling this
// will panic.
func (kv KeyValue) MustUnmarshalValueJSON(v any) {
	if err := json.Unmarshal(kv.base.Value, v); err != nil {
		panic(fmt.Errorf("failed to unmarshal KV value as JSON: %w", err))
	}
}

// UnmarshalValueYAML parses the YAML-encoded data of the KeyValue and stores the
// result in the value pointed to by v. If v is nil or not a pointer, UnmarshalValueYAML
// returns an error.
func (kv KeyValue) UnmarshalValueYAML(v any) error {
	return yaml.Unmarshal(kv.base.Value, v)
}

// MustUnmarshalValueYAML parses the YAML-encoded data of the KeyValue and stores the
// result in the value pointed to by v. If an error occurs during unmarshalling this
// will panic.
func (kv KeyValue) MustUnmarshalValueYAML(v any) {
	if err := yaml.Unmarshal(kv.base.Value, v); err != nil {
		panic(fmt.Errorf("failed to unmarshal KV value as YAML: %w", err))
	}
}

// Unwrap returns the underlying KVPair
func (kv KeyValue) Unwrap() *api.KVPair {
	return kv.base
}

// KVClient is an opinionated wrapper around the official Consul API Client for
// working with KVs in Consul.
//
// The zero-value of KVClient is not usable. Use NewKVClient to create and
// initialize a new instance of KVClient.
type KVClient struct {
	client *api.Client
}

// NewKVClient creates and initializes a new KVClient
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
func (c KVClient) Get(key string, allowStale bool) (KeyValue, error) {
	kv, _, err := c.client.KV().Get(key, &api.QueryOptions{
		AllowStale: allowStale,
	})
	// Error communicating with Consul
	if err != nil {
		return KeyValue{}, err
	}
	// Key doesn't exist
	if kv == nil {
		return KeyValue{}, nil
	}
	return KeyValue{
		base: kv,
	}, nil
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

// Put sets a value for a provided key in Consul KV store. If the operation fails
// a non-nil error value is returned.
func (c KVClient) Put(key string, value []byte) error {
	kv := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := c.client.KV().Put(kv, nil)
	return err
}

// MustPut sets a value for a provided key in Consul KV store. If the operation
// fails this will panic.
func (c KVClient) MustPut(key string, value []byte) {
	kv := &api.KVPair{
		Key:   key,
		Value: value,
	}
	if _, err := c.client.KV().Put(kv, nil); err != nil {
		panic(fmt.Errorf("failed to put KV with key %s in Consul: %w", key, err))
	}
}

// PutJSON marshals the provided value as JSON and sets that value for the given
// key in Consul KV store. If marshaling fails or putting the value in consul
// fails this returns a non-nil error value.
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

// MustPutJSON marshals the provided value as JSON and sets that value for the
// given key in Consul KV store. If an error occurs during this operation this
// will panic.
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

// PutYAML marshals the provided value as YAML and sets that value for the given
// key in Consul KV store. If marshaling fails or putting the value in consul
// fails this returns a non-nil error value.
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

// MustPutYAML marshals the provided value as YAML and sets that value for the
// given key in Consul KV store. If an error occurs during this operation this
// will panic.
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

// Delete removes a key/value from the Consul KV store. If this operation fails
// a non-nil error value is returned.
func (c KVClient) Delete(key string) error {
	_, err := c.client.KV().Delete(key, nil)
	return err
}
