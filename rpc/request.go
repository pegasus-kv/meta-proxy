package rpc

import (
	"fmt"

	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
)

// methodToRequestMap is a mapping from RPC method name to the creator the request.
var methodToRequestMap = map[string]func() rpcRequestArgs{
	"RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX": func() rpcRequestArgs {
		return &rrdb.MetaQueryCfgArgs{
			Query: replication.NewQueryCfgRequest(),
		}
	},
}

// newRPCRequestArgs creates a rpc request if it was registered.
func newRPCRequestArgs(name string) (rpcRequestArgs, error) {
	if reqCreator, ok := methodToRequestMap[name]; ok {
		return reqCreator(), nil
	} else {
		return nil, fmt.Errorf("unsupported rpc name \"%s\"", name)
	}
}
