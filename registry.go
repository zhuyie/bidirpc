package bidirpc

import "net/rpc"

// Registry provides the available RPC receivers
type Registry struct {
	server *rpc.Server
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//  - exported method of exported type
//  - two arguments, both of exported type
//  - the second argument is a pointer
//  - one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
func (r *Registry) Register(rcvr interface{}) error {
	return r.server.Register(rcvr)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (r *Registry) RegisterName(name string, rcvr interface{}) error {
	return r.server.RegisterName(name, rcvr)
}

// NewRegistry instantiates a Registry
func NewRegistry() *Registry {
	return &Registry{server: rpc.NewServer()}
}
