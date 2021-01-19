package rpc

import (
	"bytes"
	"testing"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/rpc"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/stretchr/testify/assert"
)

// TestEncoderWriteResponse ensures a response written by responseEncoder can be read by pegasus-go-client.
func TestEncoderWriteResponse(t *testing.T) {
	res := &rrdb.MetaQueryCfgResult{
		Success: &replication.QueryCfgResponse{
			Err:        &base.ErrorCode{Errno: base.ERR_INVALID_STATE.String()},
			Partitions: []*replication.PartitionConfiguration{},
		},
	}

	req := &pegasusRequest{
		seqID:      1,
		methodName: "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX",
	}
	tmpbuf := bytes.NewBuffer(nil)
	wbuf := bytes.NewBuffer(nil)
	enc := responseEncoder{
		writer: wbuf,
	}

	// the response will be encoded to wbuf
	err := enc.sendResponse(req, res)
	assert.Nil(t, err)

	// read response via pegasus-go-client response reader.
	rcall, err := session.ReadRpcResponse(rpc.NewFakeRpcConn(wbuf /*for read*/, tmpbuf), session.NewPegasusCodec())
	assert.Nil(t, err)
	queryCfgRes, ok := rcall.Result.(*rrdb.MetaQueryCfgResult)
	assert.True(t, ok)
	assert.Equal(t, *res.Success, *queryCfgRes.Success)
}
