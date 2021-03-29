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
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

type requestV0 struct {
	hdrLength uint32 // always 48

	meta *requestMetaV0
}

type requestMetaV0 struct {
	/* hdrCRC32 int32 */ // unused
	bodyLength           uint32
	/* bodyCRC32 uint32 */ // unused
	appID                  uint32
	partitionIndex         uint32
	clientTimeout          uint32
	clientThreadHash       uint32
	clientPartitionHash    uint64
}

type requestV1 struct {
	metaLength uint32
	bodyLength uint32

	meta *ThriftRequestMetaV1
}

type requestDecoder struct {
	reader io.Reader
}

// pegasusProtocolFlag is a const flag to identify if the RPC request is legal.
var pegasusProtocolFlag = []byte("THFT")

//
// For version 0:
// |<--              fixed-size request header              -->|<--request body-->|
// |-"THFT"-|- hdr_version + hdr_length -|-  request_meta_v0  -|-      blob      -|
// |-"THFT"-|-  uint32(0)  + uint32(48) -|-      36bytes      -|-                -|
// |-               12bytes             -|-      36bytes      -|-                -|
//
// For version 1:
// |<--          fixed-size request header           -->| <--        request body        -->|
// |-"THFT"-|- hdr_version + meta_length + body_length -|- thrift_request_meta_v1 -|- blob -|
// |-"THFT"-|-  uint32(1)  +   uint32    +    uint32   -|-      thrift struct     -|-      -|
// |-                      16bytes                     -|-      thrift struct     -|-      -|
//

type pegasusRequest struct {
	// reqv0 / reqv1 can have only one to be non-nil.
	reqv0 *requestV0
	reqv1 *requestV1

	methodName string
	seqID      uint64
	args       RequestArgs
	handler    MethodHandler
}

// readRequest reads fully the RPC request into pegasusRequest.
func (d *requestDecoder) readRequest() (*pegasusRequest, error) {
	// read protocol flag
	flag := make([]byte, 4)
	_, err := io.ReadFull(d.reader, flag)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(flag, pegasusProtocolFlag) {
		return nil, fmt.Errorf("invalid rdsn rpc protocol: %s", flag)
	}

	// read header version
	hdrVersionBytes := make([]byte, 4)
	_, err = io.ReadFull(d.reader, hdrVersionBytes)
	if err != nil {
		return nil, err
	}
	hdrVersion := binary.BigEndian.Uint32(hdrVersionBytes)
	if hdrVersion == 0 {
		return d.readRequestV0()
	} else if hdrVersion == 1 {
		return d.readRequestV1()
	}
	return nil, fmt.Errorf("invalid request header version: %d", hdrVersion)
}

func (d *requestDecoder) readRequestV0() (*pegasusRequest, error) {
	// |- hdr_length -|-  request_meta_v0  -|
	// |- uint32(48) -|-      36bytes      -|
	// |-           40bytes                -|

	data := make([]byte, 40)
	_, err := io.ReadFull(d.reader, data)
	if err != nil {
		return nil, err
	}

	// read header length
	pegasusReq := &pegasusRequest{reqv0: &requestV0{}}
	reqv0 := pegasusReq.reqv0
	hdrLength := binary.BigEndian.Uint32(data[0:4])
	if hdrLength != 48 {
		return nil, fmt.Errorf("header length (%d) is not 48", hdrLength)
	}
	reqv0.hdrLength = hdrLength

	// read request meta
	reqMeta := &requestMetaV0{}
	reqMeta.bodyLength = binary.BigEndian.Uint32(data[8:12])
	reqMeta.appID = binary.BigEndian.Uint32(data[16:20])
	reqMeta.partitionIndex = binary.BigEndian.Uint32(data[20:24])
	reqMeta.clientTimeout = binary.BigEndian.Uint32(data[24:28])
	reqMeta.clientThreadHash = binary.BigEndian.Uint32(data[28:32])
	reqMeta.clientPartitionHash = binary.BigEndian.Uint64(data[32:40])
	reqv0.meta = reqMeta

	// read request body
	err = d.readRequestBody(pegasusReq, reqMeta.bodyLength)
	if err != nil {
		return nil, err
	}

	return pegasusReq, nil
}

func (d *requestDecoder) readRequestV1() (*pegasusRequest, error) {
	//	|- meta_length + body_length -|- thrift_request_meta_v1 -|- blob -|
	//	|-   uint32    +    uint32   -|-      thrift struct     -|-      -|

	data := make([]byte, 8)
	_, err := io.ReadFull(d.reader, data)
	if err != nil {
		return nil, err
	}

	pegasusReq := &pegasusRequest{reqv1: &requestV1{}}
	reqv1 := pegasusReq.reqv1

	reqv1.metaLength = binary.BigEndian.Uint32(data[0:4])
	reqv1.bodyLength = binary.BigEndian.Uint32(data[4:8])

	// read thrift_request_meta_v1
	// TODO(wutao): do we need this struct?
	data = make([]byte, reqv1.metaLength)
	_, err = io.ReadFull(d.reader, data)
	if err != nil {
		return nil, err
	}
	iprot := thrift.NewTBinaryProtocolTransport(thrift.NewStreamTransportR(bytes.NewBuffer(data)))
	meta := NewThriftRequestMetaV1()
	err = meta.Read(iprot)
	if err != nil {
		return nil, err
	}
	reqv1.meta = meta

	// read request body
	err = d.readRequestBody(pegasusReq, reqv1.bodyLength)
	if err != nil {
		return nil, err
	}

	return pegasusReq, nil
}

// The request body encoding is common in both v0/v1 RPC protocol.
func (d *requestDecoder) readRequestBody(req *pegasusRequest, bodyLength uint32) error {
	data := make([]byte, bodyLength)
	_, err := io.ReadFull(d.reader, data)
	if err != nil {
		return err
	}

	iprot := thrift.NewTBinaryProtocolTransport(thrift.NewStreamTransportR(bytes.NewBuffer(data)))

	name, _, seq, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}
	req.seqID = uint64(seq)
	req.methodName = name
	method, err := findMethodByName(name)
	if err != nil {
		return err
	}
	req.handler = method.Handler
	req.args = method.RequestCreator()
	if err = req.args.Read(iprot); err != nil {
		return err
	}
	if err = iprot.ReadMessageEnd(); err != nil {
		return err
	}
	return nil
}
