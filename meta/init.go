package meta

import (
	"context"
	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/pegasus-kv/meta-proxy/collector"
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
		Handler: queryConfig,
	})
}

func queryConfig(ctx context.Context, args rpc.RequestArgs) rpc.ResponseResult {
	collector.ClientQueryConfigQPS.Update()

	var errorCode *base.ErrorCode
	queryCfgArgs := args.(*rrdb.MetaQueryCfgArgs)
	tableName := queryCfgArgs.Query.AppName
	meta, err := globalClusterManager.getMeta(tableName)
	if err != nil {
		errorCode = parseToErrorCode(err)
		return &rrdb.MetaQueryCfgResult{
			Success: &replication.QueryCfgResponse{
				Err: errorCode,
			},
		}
	}

	resp, err := meta.QueryConfig(ctx, tableName)
	if err != nil {
		errorCode = parseToErrorCode(err)
		return &rrdb.MetaQueryCfgResult{
			Success: &replication.QueryCfgResponse{
				Err: errorCode,
			},
		}
	}

	return &rrdb.MetaQueryCfgResult{
		Success: resp,
	}
}

func parseToErrorCode(err error) *base.ErrorCode {
	if dsnErr, ok := err.(base.DsnErrCode); ok {
		return &base.ErrorCode{Errno: dsnErr.String()}
	} else {
		return &base.ErrorCode{Errno: base.ERR_UNKNOWN.String()}
	}
}
