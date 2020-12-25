package meta

import (
	"context"
	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/pegasus-kv/meta-proxy/rpc"
)

func Init() {
	initClusterManager()

	rpc.Register("RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX", &rpc.MethodDefinition{
		RequestCreator: func() rpc.RequestArgs {
			return &rrdb.MetaQueryCfgArgs{
				Query: replication.NewQueryCfgRequest(),
			}
		},
		Handler: globalClusterManager.queryConfig,
	})
}

func (m *ClusterManager) queryConfig(ctx context.Context, args rpc.RequestArgs) rpc.ResponseResult {
	var dsnErrorCode base.DsnErrCode

	queryCfgArgs := args.(*rrdb.MetaQueryCfgArgs)
	tableName := queryCfgArgs.Query.AppName
	meta, err := m.getMetaConnector(tableName)
	if err != nil {
		dsnErrorCode = parseToDsnErrCode(err)
		return &rrdb.MetaQueryCfgResult{
			Success: &replication.QueryCfgResponse{
				Err: &base.ErrorCode{Errno: dsnErrorCode.String()},
			},
		}
	}

	resp, err := meta.QueryConfig(ctx, tableName)
	if err != nil {
		dsnErrorCode = parseToDsnErrCode(err)
		return &rrdb.MetaQueryCfgResult{
			Success: &replication.QueryCfgResponse{
				Err: &base.ErrorCode{Errno: dsnErrorCode.String()},
			},
		}
	}

	return &rrdb.MetaQueryCfgResult{
		Success: resp,
	}
}

func parseToDsnErrCode(err error) base.DsnErrCode {
	if dsnErr, ok := err.(base.DsnErrCode); ok {
		return dsnErr
	} else {
		return base.ERR_UNKNOWN
	}
}
