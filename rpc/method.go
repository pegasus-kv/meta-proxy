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
	"fmt"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

func init() {
	globalMethodRegistry.nameToMethod = make(map[string]*MethodDefinition)
}

// methodRegistry stores the mapping from RPC method name to the method definition.
type methodRegistry struct {
	nameToMethod map[string]*MethodDefinition
}

func findMethodByName(name string) (*MethodDefinition, error) {
	if method, ok := (&globalMethodRegistry).nameToMethod[name]; ok {
		return method, nil
	}
	return nil, fmt.Errorf("unsupported rpc name \"%s\"", name)
}

var globalMethodRegistry methodRegistry

// Register an RPC method.
func Register(name string, method *MethodDefinition) {
	globalMethodRegistry.nameToMethod[name] = method
}

// MethodHandler handles a rpc request
type MethodHandler func(context.Context, RequestArgs) ResponseResult

// MethodDefinition defines the RPC method.
type MethodDefinition struct {
	RequestCreator func() RequestArgs

	Handler MethodHandler
}

// RequestArgs is any type of request.
type RequestArgs interface {
	String() string
	Read(iprot thrift.TProtocol) error
}

// ResponseResult is any type of response.
type ResponseResult interface {
	String() string
	Write(oprot thrift.TProtocol) error
}
