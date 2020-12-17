package rpc

import (
	"bytes"
	"context"
	"testing"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/stretchr/testify/assert"
)

// fakeConn implements interface io.ReadWriteCloser
type fakeConn struct {
	*bytes.Buffer
}

func (*fakeConn) Close() error {
	return nil
}

func newFakeConn(bs []byte) *fakeConn {
	return &fakeConn{Buffer: bytes.NewBuffer(bs)}
}

func TestDecoderReadRequest(t *testing.T) {
	seqID := int32(1)
	gpid := &base.Gpid{Appid: 3, PartitionIndex: 4}
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()
	arg.Query.AppName = "test"
	arg.Query.PartitionIndices = []int32{}

	// register method
	Register("RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX", &MethodDefinition{
		RequestCreator: func() RequestArgs {
			return &rrdb.MetaQueryCfgArgs{
				Query: replication.NewQueryCfgRequest(),
			}
		},
	})

	rcall, err := session.MarshallPegasusRpc(session.NewPegasusCodec(), seqID, gpid, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)

	dec := &requestDecoder{
		reader: newFakeConn(rcall.RawReq),
	}
	req, err := dec.readRequest()
	assert.Nil(t, err)
	assert.Equal(t, req.seqID, uint64(seqID))
	assert.Equal(t, req.reqv0.meta.appID, uint32(gpid.Appid))
	assert.Equal(t, req.reqv0.meta.partitionIndex, uint32(gpid.PartitionIndex))

	queryCfgArg, ok := req.args.(*rrdb.MetaQueryCfgArgs)
	assert.True(t, ok)
	assert.Equal(t, *queryCfgArg, *arg)

	// do cleanup after test
	globalMethodRegistry.nameToMethod = make(map[string]*MethodDefinition)
}

// TestDecoderHandleRequest ensures a request can invokes its corresponding method.
func TestDecoderHandleRequest(t *testing.T) {
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()

	Register("RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX", &MethodDefinition{
		RequestCreator: func() RequestArgs {
			return &rrdb.MetaQueryCfgArgs{
				Query: replication.NewQueryCfgRequest(),
			}
		},
		Handler: func(c context.Context, ra RequestArgs) ResponseResult {
			return &rrdb.MetaQueryCfgResult{
				Success: &replication.QueryCfgResponse{
					Err: &base.ErrorCode{Errno: base.ERR_INVALID_STATE.String()},
				},
			}
		},
	})

	rcall, err := session.MarshallPegasusRpc(session.NewPegasusCodec(), int32(1), &base.Gpid{}, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)
	dec := &requestDecoder{reader: newFakeConn(rcall.RawReq)}
	req, err := dec.readRequest()
	assert.Nil(t, err)

	resp := req.handler(context.Background(), req.args)
	queryCfgResp := resp.(*rrdb.MetaQueryCfgResult)
	assert.Equal(t, queryCfgResp.Success.Err.Errno, "ERR_INVALID_STATE")

	// do cleanup after test
	globalMethodRegistry.nameToMethod = make(map[string]*MethodDefinition)
}

func TestDecoderHandleUnsupportedRequest(t *testing.T) {
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()

	rcall, err := session.MarshallPegasusRpc(session.NewPegasusCodec(), int32(1), &base.Gpid{}, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)
	dec := &requestDecoder{reader: newFakeConn(rcall.RawReq)}
	_, err = dec.readRequest()
	assert.NotNil(t, err) // method-not-found
}
