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
	"encoding/binary"
	"io"
	"sync"

	"github.com/apache/thrift/lib/go/thrift"
)

type responseEncoder struct {
	writer io.Writer

	mu sync.Mutex
}

// sendResponse is a thread-safe wrapper of doSendResponse.
func (e *responseEncoder) sendResponse(req *pegasusRequest, result ResponseResult) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.doSendResponse(req, result)
}

func (e *responseEncoder) doSendResponse(req *pegasusRequest, result ResponseResult) error {
	// prepare response bytes

	buf := thrift.NewTMemoryBuffer()
	oprot := thrift.NewTBinaryProtocolTransport(buf)
	err := oprot.WriteI32(0) // response length is still unknown now.
	if err != nil {
		return err
	}

	// error code
	if err = oprot.WriteString("ERR_OK"); err != nil {
		return err
	}

	// write response
	if err = oprot.WriteMessageBegin(req.methodName+"_ACK", thrift.REPLY, int32(req.seqID)); err != nil {
		return err
	}
	if err = result.Write(oprot); err != nil {
		return err
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return err
	}

	// response length is now got
	respLen := buf.Len()
	binary.BigEndian.PutUint32(buf.Bytes(), uint32(respLen))

	if _, err := e.writer.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
