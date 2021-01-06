package rpc

import (
	"github.com/pegasus-kv/meta-proxy/collector"
	"github.com/pegasus-kv/meta-proxy/config"
	"sync"
	"testing"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/rpc"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.InitConfig("../meta-proxy.yml")
	collector.InitPerfCounter()
}

func TestServeConn(t *testing.T) {
	// mock connection and request
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = &replication.QueryCfgRequest{
		AppName:          "test",
		PartitionIndices: []int32{},
	}

	rcall, _ := session.MarshallPegasusRpc(session.NewPegasusCodec(), int32(1), &base.Gpid{}, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	reqCnt := 50 // fill the connection with `reqCnt` requests to test robustness under concurrent case
	var reqBuf []byte
	for i := 0; i < reqCnt; i++ {
		reqBuf = append(reqBuf, rcall.RawReq...)
	}
	conn := newFakeConn(reqBuf)

	// QueryConfig always returns a successful response
	resp := &replication.QueryCfgResponse{
		Err:            &base.ErrorCode{Errno: "ERR_OK"},
		AppID:          3,
		PartitionCount: 128,
		Partitions:     []*replication.PartitionConfiguration{},
	}
	testSetUpQueryConfigRPC(resp)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		serveConn(conn, "127.0.0.1:56789")
		wg.Done()
	}()
	wg.Wait()

	// conn.wbuf is the response bytes from server
	assert.Greater(t, conn.wbuf.Len(), 0)
	clientConn := rpc.NewFakeRpcConn(conn.wbuf, nil)
	for i := 0; i < reqCnt; i++ {
		rcall, err := session.ReadRpcResponse(clientConn, session.NewPegasusCodec())
		assert.Nil(t, err)
		queryCfgRes, ok := rcall.Result.(*rrdb.MetaQueryCfgResult)
		assert.True(t, ok)
		assert.Equal(t, *queryCfgRes.Success, *resp)
	}

	testCleanupRPCRegsitration()
}
