package meta

import (
	"context"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	initClusterManager()
}

func TestQueryConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	args := &rrdb.MetaQueryCfgArgs{
		Query: replication.NewQueryCfgRequest(),
	}
	args.Query.AppName = "temp"
	resp := clusterManager.queryConfig(ctx, args).(*rrdb.MetaQueryCfgResult)
	assert.Equal(t, len(resp.Success.Partitions), 8)
}
