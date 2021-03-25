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
	"testing"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/stretchr/testify/assert"
)

func init() {
	Init()
}

func TestQueryConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	args := &rrdb.MetaQueryCfgArgs{
		Query: replication.NewQueryCfgRequest(),
	}
	args.Query.AppName = "temp"
	resp := queryConfig(ctx, args).(*rrdb.MetaQueryCfgResult)
	assert.Equal(t, 8, len(resp.Success.Partitions))
	assert.Equal(t, &base.ErrorCode{Errno: base.ERR_OK.String()}, resp.Success.Err)

	args.Query.AppName = "notExist"
	resp = queryConfig(ctx, args).(*rrdb.MetaQueryCfgResult)
	assert.Equal(t, &base.ErrorCode{Errno: base.ERR_OBJECT_NOT_FOUND.String()}, resp.Success.Err)
}
