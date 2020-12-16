package rpc

import (
	"bytes"
	"testing"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/stretchr/testify/assert"
)

type fakeConn struct {
	*bytes.Buffer
}

func (*fakeConn) Close() error {
	return nil
}

func newFakeConn(bs []byte) *fakeConn {
	return &fakeConn{Buffer: bytes.NewBuffer(bs)}
}

func TestServeCodec(t *testing.T) {
	seqID := int32(1)
	gpid := &base.Gpid{}
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()
	arg.Query.AppName = "test"
	arg.Query.PartitionIndices = []int32{}

	rcall, err := session.MarshallPegasusRpc(session.NewPegasusCodec(), seqID, gpid, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)

	serveConn(newFakeConn(rcall.RawReq))
}
