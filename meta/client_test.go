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

package meta

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/go-zookeeper/zk"
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"
)

type testCase struct {
	table string
	addr  string
	path  string
	data  string
}

var zkRootTest = "/pegasus-cluster"
var tests = []testCase{
	{
		table: "temp",
		addr:  "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		path:  zkRootTest + "/temp",
		data:  "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}"},
	{
		table: "stat",
		addr:  "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		path:  zkRootTest + "/stat",
		data:  "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}"},
	{
		table: "test",
		addr:  "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		path:  zkRootTest + "/test",
		data:  "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}"},
}

type update struct {
	addr string
	data string
}

var updates = []update{
	{
		addr: "127.1.1.1:34601,127.1.1.1:34602,127.1.1.1:34603",
		data: "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.1.1.1:34601,127.1.1.1:34602,127.1.1.1:34603\"}",
	},
	{
		addr: "127.0.1.1:34601,127.0.1.1:34602,127.0.1.1:34603",
		data: "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.1.1:34601,127.0.1.1:34602,127.0.1.1:34603\"}",
	},
	{
		addr: "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		data: "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}",
	},
}

func initTestLog() {
	writers := []io.Writer{
		&lumberjack.Logger{
			Filename:  "meta-proxy-test.log",
			LocalTime: true,
		},
		os.Stdout}
	logrus.SetOutput(io.MultiWriter(writers...))
}

// init the zk data
func init() {
	initTestLog()
	config.Init("../config/yaml/meta-proxy-example.yml")
	config.GlobalConfig.ZookeeperOpts.WatcherCount = 2
	initClusterManager()

	acls := zk.WorldACL(zk.PermAll)
	zkRoot := config.GlobalConfig.ZookeeperOpts.Root
	ret, _, _ := globalClusterManager.ZkConn.Exists(zkRoot)
	if !ret {
		_, err := globalClusterManager.ZkConn.Create(zkRoot, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			panic(err)
		}
	}

	for _, test := range tests {
		ret, stat, _ := globalClusterManager.ZkConn.Exists(test.path)
		if ret {
			_ = globalClusterManager.ZkConn.Delete(test.path, stat.Version)
		}
		_, err := globalClusterManager.ZkConn.Create(test.path, []byte(test.data), 0, acls)
		if err != nil {
			panic(err)
		}
	}
}

func TestGetTable(t *testing.T) {
	// pass zkAddr can't be connected
	config.GlobalConfig.ZookeeperOpts.Address = []string{"128.0.0.1:22171"}
	initClusterManager()
	_, err := globalClusterManager.newTableInfo("notExist")
	assert.Equal(t, err, base.ERR_ZOOKEEPER_OPERATION)

	config.GlobalConfig.ZookeeperOpts.Address = []string{"127.0.0.1:22181"}
	initClusterManager()
	// pass not existed table name
	_, err = globalClusterManager.newTableInfo("notExist")
	assert.Equal(t, err, base.ERR_OBJECT_NOT_FOUND)
	// pass exist table
	for _, test := range tests {
		addrs, err := globalClusterManager.newTableInfo(test.table)
		if err != nil {
			logrus.Panic(err)
		}
		assert.Equal(t, test.addr, addrs.metaAddrs)
	}
}

func TestGetMetaConnector(t *testing.T) {
	config.GlobalConfig.ZookeeperOpts.Address = []string{"127.0.0.1:22181"}
	initClusterManager()

	// first get connector which will init the cache and only store `stat` and `test` table watcher
	for _, test := range tests {
		_, _, _ = globalClusterManager.getMeta(test.table)
		cacheWatcher, _ := globalClusterManager.Tables.Get(test.table)
		assert.Equal(t, test.addr, cacheWatcher.(*TableInfoWatcher).metaAddrs)
	}

	// zkWatcherCount = 2
	assert.Equal(t, globalClusterManager.Tables.Len(true), 2)
	// second get connector which will be get from cache
	for _, test := range tests {
		cacheWatcher, _ := globalClusterManager.Tables.Get(test.table)
		// zkWatcherCount = 2, the oldest cache "temp" is removed
		if test.table == "temp" {
			assert.Nil(t, cacheWatcher)
		} else {
			assert.Equal(t, test.addr, cacheWatcher.(*TableInfoWatcher).metaAddrs)
			assert.NotNil(t, globalClusterManager.Metas[test.addr])
			_, meta, _ := globalClusterManager.getMeta(test.table)
			assert.NotNil(t, meta)
		}
	}
}

func TestZookeeperUpdate(t *testing.T) {
	zkRoot := config.GlobalConfig.ZookeeperOpts.Root
	for _, test := range tests {
		_, _, _ = globalClusterManager.getMeta(test.table)
		// update zookeeper node data and trigger the watch event update local cache
		for _, update := range updates {
			_, stat, _ := globalClusterManager.ZkConn.Get(test.path)
			_, err := globalClusterManager.ZkConn.Set(test.path, []byte(update.data), stat.Version)
			if err != nil {
				panic(err)
			}

			// local cache will change to new meta addr because the zk watcher
			time.Sleep(time.Duration(100000000))
			cacheWatcher, _ := globalClusterManager.Tables.Get(test.table)
			assert.Equal(t, update.addr, cacheWatcher.(*TableInfoWatcher).metaAddrs)
		}
	}

	// delete the zk node data and the local local cache is also removed
	_, stat, _ := globalClusterManager.ZkConn.Get(zkRoot + "/test")
	_ = globalClusterManager.ZkConn.Delete(zkRoot+"/test", stat.Version)
	time.Sleep(time.Duration(10000000))
	cacheWatcher, _ := globalClusterManager.Tables.Get("test")
	assert.Nil(t, cacheWatcher)
}

func TestParseTablePath(t *testing.T) {
	zkRoot := config.GlobalConfig.ZookeeperOpts.Root
	type table struct {
		path   string
		result error
	}

	tablePaths := []table{
		{
			path:   zkRoot,
			result: fmt.Errorf("the path[%s] is invalid", zkRoot),
		},
		{
			path:   zkRoot + "/name",
			result: nil,
		},
		{
			path:   zkRoot + "//name",
			result: nil,
		},
		{
			path:   zkRoot + "/table/name",
			result: nil,
		},
	}

	for _, tb := range tablePaths {
		ret, err := parseToTableName(tb.path)
		assert.Equal(t, err, tb.result)
		if err == nil {
			assert.Equal(t, "name", ret)
		}
	}
}

func TestParseMetaAddrs(t *testing.T) {
	type meta struct {
		addrs  string
		result error
	}

	metas := []meta{
		{
			addrs:  "127.0.0.1,127.0.0.2",
			result: nil,
		},
		{
			addrs:  "127.0.0.1",
			result: fmt.Errorf("the meta addrs[%s] is invalid", "127.0.0.1"),
		},
	}

	for _, m := range metas {
		ret, err := parseToMetaList(m.addrs)
		assert.Equal(t, err, m.result)
		if err == nil {
			assert.Equal(t, ret, []string{"127.0.0.1", "127.0.0.2"})
		}
	}
}
