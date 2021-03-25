package rpc

import (
	"context"
	"fmt"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

func init() {
	globalMethodRegistry.nameToMethod = make(map[string]*MethodDefinition)
}

// methodRegistry stores the mapping from RPC method name to the method definition.
type methodRegistry struct {
	nameToMethod map[string]*MethodDefinition
}

func findMethodByName(name string) (*MethodDefinition, error) {
	if method, ok := (&globalMethodRegistry).nameToMethod[name]; ok {
		return method, nil
	}
	return nil, fmt.Errorf("unsupported rpc name \"%s\"", name)
}

var globalMethodRegistry methodRegistry

// Register an RPC method.
func Register(name string, method *MethodDefinition) {
	globalMethodRegistry.nameToMethod[name] = method
}

// MethodHandler handles a rpc request
type MethodHandler func(context.Context, RequestArgs) ResponseResult

// MethodDefinition defines the RPC method.
type MethodDefinition struct {
	RequestCreator func() RequestArgs

	Handler MethodHandler
}

// RequestArgs is any type of request.
type RequestArgs interface {
	String() string
	Read(iprot thrift.TProtocol) error
}

// ResponseResult is any type of response.
type ResponseResult interface {
	String() string
	Write(oprot thrift.TProtocol) error
}
