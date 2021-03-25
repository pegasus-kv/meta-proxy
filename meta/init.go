/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package meta

import (
	"context"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/pegasus-kv/meta-proxy/metrics"
	"github.com/pegasus-kv/meta-proxy/rpc"
)

var clientQueryConfigQPS metrics.Meter

func Init() {
	clientQueryConfigQPS = metrics.RegisterMeter("client_query_config_qps")
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
	clientQueryConfigQPS.Update()

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
	}
	return &base.ErrorCode{Errno: base.ERR_UNKNOWN.String()}
}
