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

package rpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"net/http"
	"testing"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/stretchr/testify/assert"
)

// fakeConn implements interface io.ReadWriteCloser
type fakeConn struct {
	rbuf, wbuf *bytes.Buffer
}

func (*fakeConn) Close() error {
	return nil
}

func (f *fakeConn) Read(p []byte) (int, error) {
	return f.rbuf.Read(p)
}

func (f *fakeConn) Write(p []byte) (int, error) {
	return f.wbuf.Write(p)
}

func newFakeConn(readBytes []byte) *fakeConn {
	return &fakeConn{rbuf: bytes.NewBuffer(readBytes), wbuf: bytes.NewBuffer(nil)}
}

func registerQueryConfigRPC(resp *replication.QueryCfgResponse) {
	Register("RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX", &MethodDefinition{
		RequestCreator: func() RequestArgs {
			return &rrdb.MetaQueryCfgArgs{
				Query: replication.NewQueryCfgRequest(),
			}
		},
		Handler: func(c context.Context, ra RequestArgs) ResponseResult {
			return &rrdb.MetaQueryCfgResult{
				Success: resp,
			}
		},
	})
}

func unregisterAllRPC() {
	// do cleanup after test
	globalMethodRegistry.nameToMethod = make(map[string]*MethodDefinition)
}

func TestDecoderReadRequest(t *testing.T) {
	seqID := int32(1)
	gpid := &base.Gpid{Appid: 3, PartitionIndex: 4}
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()
	arg.Query.AppName = "test"
	arg.Query.PartitionIndices = []int32{}

	// register method
	registerQueryConfigRPC(nil)
	defer unregisterAllRPC()

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

}

// TestDecoderHandleRequest ensures a request can invokes its corresponding method.
func TestDecoderHandleRequest(t *testing.T) {
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()

	// QueryConfig definitely returns ERR_INVALID_STATE
	registerQueryConfigRPC(&replication.QueryCfgResponse{
		Err: &base.ErrorCode{Errno: base.ERR_INVALID_STATE.String()},
	})
	defer unregisterAllRPC()

	rcall, err := session.MarshallPegasusRpc(session.NewPegasusCodec(), int32(1), &base.Gpid{}, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)
	dec := &requestDecoder{reader: newFakeConn(rcall.RawReq)}
	req, err := dec.readRequest()
	assert.Nil(t, err)

	resp := req.handler(context.Background(), req.args)
	queryCfgResp := resp.(*rrdb.MetaQueryCfgResult)
	assert.Equal(t, queryCfgResp.Success.Err.Errno, "ERR_INVALID_STATE")
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

func TestUnexpectedRPCProtocol(t *testing.T) {
	body := bytes.NewBufferString("hello")
	httpReq, err := http.NewRequest(http.MethodGet, "www.baidu.com", body)
	assert.Nil(t, err)

	// marshall a fake http request into bytes.
	buf := bytes.NewBuffer(nil)
	err = httpReq.Write(buf)
	assert.Nil(t, err)

	// verify if requestDecoder fails on receiving invalid rpc protocol.
	dec := &requestDecoder{reader: newFakeConn(buf.Bytes())}
	_, err = dec.readRequest()
	assert.NotNil(t, err)
}

func TestV1ProtocolReadRequest(t *testing.T) {
	seqID := int32(1)
	gpid := &base.Gpid{Appid: 3, PartitionIndex: 4}
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()
	arg.Query.AppName = "test"
	arg.Query.PartitionIndices = []int32{}

	// register method
	registerQueryConfigRPC(nil)
	defer unregisterAllRPC()

	rcall, err := session.MarshallPegasusRpc(&pegasusV1Codec{}, seqID, gpid, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)

	dec := &requestDecoder{
		reader: newFakeConn(rcall.RawReq),
	}
	req, err := dec.readRequest()
	assert.Nil(t, err)
	assert.Equal(t, req.seqID, uint64(seqID))
	assert.Equal(t, req.reqv1.meta.GetAppID(), gpid.Appid)
	assert.Equal(t, req.reqv1.meta.GetPartitionIndex(), gpid.PartitionIndex)
}

func TestV1ProtocolHandleBadRequest(t *testing.T) {
	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()

	rcall, err := session.MarshallPegasusRpc(&pegasusV1Codec{}, int32(1), &base.Gpid{}, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	assert.Nil(t, err)
	binary.BigEndian.PutUint32(rcall.RawReq[4:8], 999) // illegal protocol version

	dec := &requestDecoder{reader: newFakeConn(rcall.RawReq)}
	_, err = dec.readRequest()
	assert.NotNil(t, err) // invalid request header version

	for i := 1; i < len(rcall.RawReq); i++ {
		buf := rcall.RawReq[:i] // truncate a part of the request, see if our error handling is correct

		dec := &requestDecoder{reader: newFakeConn(buf)}
		_, err = dec.readRequest()
		assert.NotNil(t, err)
	}
}
