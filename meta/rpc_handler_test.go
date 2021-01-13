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
