package rpc

import (
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

// methodRegistry stores the mapping from RPC method name to the method definition.
type methodRegistry struct {
	nameToMethod map[string]*MethodDefinition
}

func findMethodByName(name string) (*MethodDefinition, error) {
	if method, ok := (&globalMethodRegistry).nameToMethod[name]; ok {
		return method, nil
	} else {
		return nil, fmt.Errorf("unsupported rpc name \"%s\"", name)
	}
}

var globalMethodRegistry methodRegistry

// Register an RPC method.
func Register(name string, method *MethodDefinition) {
	globalMethodRegistry.nameToMethod[name] = method
}

type MethodHandler func(RequestArgs) ResponseResult

// MethodDefinition defines the RPC method.
type MethodDefinition struct {
	RequestCreator func() RequestArgs

	Handler MethodHandler
}

type RequestArgs interface {
	String() string
	Read(iprot thrift.TProtocol) error
}

type ResponseResult interface {
	String() string
	Write(oprot thrift.TProtocol) error
}
