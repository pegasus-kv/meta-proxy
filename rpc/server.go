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
	"context"
	"io"
	"net"
	"sync"

	"github.com/pegasus-kv/meta-proxy/metrics"
	"github.com/sirupsen/logrus"
)

// declare perfcounters
var clientConnectionCount metrics.Gauge

// Serve blocks until the connection shutdown.
func Serve() error {
	clientConnectionCount = metrics.RegisterGauge("client_connection_count")

	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:34601")
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	logrus.Infof("start server listen: %s", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("connection accept: %s", err)
			continue
		}
		clientConnectionCount.Inc()
		// TODO(wutao): add connections management

		// use one goroutine per connection
		go serveConn(conn, conn.RemoteAddr().String())
	}
}

// conn is a network connection but abstracted as a ReadWriteCloser here in order to do mock test.
// The caller typically invokes serveConn in a go statement.
func serveConn(conn io.ReadWriteCloser, remoteAddr string) {
	dec := &requestDecoder{
		reader: conn,
	}
	enc := &responseEncoder{
		writer: conn,
	}

	// `ctx` is the root of all sub-tasks. It notifies the children to terminate
	//  if the connection encounters some error.
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for {
		req, err := dec.readRequest()
		if err != nil {
			if err != io.EOF {
				logrus.Warn(err)
				// TODO(wutao): send back rpc response for this error if the request is fully read
				continue
			}
			clientConnectionCount.Dec()
			logrus.Infof("connection %s is closed", remoteAddr)
			break
		}

		// Asynchronously execute RPC handler in order to not block the connection reading.
		wg.Add(1)
		go func() {
			result := req.handler(ctx, req.args)
			err := enc.sendResponse(req, result)
			if err != nil {
				logrus.Error(err)
			}

			wg.Done()
		}()
	}

	// cancel the ongoing requests
	cancel()

	// This connection exits only when all children are terminated.
	wg.Wait()
}
