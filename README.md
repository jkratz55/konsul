# Konsul

Konsul is an opinionated wrapper around the official Consul GO Client/SDK client. I've often had to work with configuration stored in Consul KVs and performing client side load balancing for services registered in Consul. I built this library to streamline working with Consul from Go with convenient wrappers and utilities.

Konsul provides the following:

* Wrapper around KV client to that streamlines handling fetching KVs and unmarshalling the values. The API includes several `Must` methods to panic on error since I've encountered many cases where if fetching configuration stored in Consul fails the application cannot start up.
* Reloadable interface and a Watch function to watch a specific KV and automatically unmarshall and reload configuration on change.
* An Instancer type to implement client side load balancing of a Consul service.
* Wrappers to allow zap and zerolog to work with Consul API. The wrappers implement the hclog.Logger interface.

There are examples that can be referenced in the examples directory.
