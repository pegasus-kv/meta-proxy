package meta

import (
	"context"
	"fmt"
	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/pegalog"
	"github.com/pegasus-kv/meta-proxy/rpc"
)

func Init() {
	rpc.Register("RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX", &rpc.MethodDefinition{
		RequestCreator: func() rpc.RequestArgs {
			return &rrdb.MetaQueryCfgArgs{
				Query: replication.NewQueryCfgRequest(),
			}
		},
		Handler: clusterManager.queryConfig,
	})
}

func (m *ClusterManager) queryConfig(ctx context.Context, args rpc.RequestArgs) rpc.ResponseResult {
	queryCfgArgs := args.(*rrdb.MetaQueryCfgArgs)
	tableName := queryCfgArgs.Query.AppName
	resp, err := m.getMetaConnector(tableName).QueryConfig(ctx, tableName)
	if err == nil {
		return &rrdb.MetaQueryCfgResult{
			Success: resp,
		}
	}

	errMsg := fmt.Sprintf("[%s]%s", tableName, err.Error())
	pegalog.GetLogger().Fatal(errMsg) // TODO(jiashuo1)
	return &rrdb.MetaQueryCfgResult{
		Success: &replication.QueryCfgResponse{
			Err: &base.ErrorCode{Errno: errMsg},
		},
	}
}
