package meta

import (
	"context"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/pegasus-kv/meta-proxy/metrics"
	"github.com/pegasus-kv/meta-proxy/rpc"
	"github.com/sirupsen/logrus"
)

var clientQueryConfigQPS metrics.Meter

func Init() {
	clientQueryConfigQPS = metrics.RegisterMeterWithTags("client_query_config_qps", []string{"table"})
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
	var errorCode *base.ErrorCode
	queryCfgArgs := args.(*rrdb.MetaQueryCfgArgs)
	tableName := queryCfgArgs.Query.AppName
	clientQueryConfigQPS.UpdateWithTags([]string{tableName})

	addrs, meta, err := globalClusterManager.getMeta(tableName)
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

	if resp.GetErr().Errno != base.ERR_OK.String() {
		logrus.Errorf("[%s] failed to query config from [%s], err = %s", tableName, addrs, resp.Err)
	}

	return &rrdb.MetaQueryCfgResult{
		Success: resp,
	}
}

func parseToErrorCode(err error) *base.ErrorCode {
	if dsnErr, ok := err.(base.DsnErrCode); ok {
		return &base.ErrorCode{Errno: dsnErr.String()}
	}
	return &base.ErrorCode{Errno: base.ERR_UNKNOWN.String()}
}
