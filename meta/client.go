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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/bluele/gcache"
	"github.com/go-zookeeper/zk"
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/pegasus-kv/meta-proxy/metrics"
	"github.com/sirupsen/logrus"
)

var zkRequestCount metrics.Meter

var globalClusterManager *ClusterManager

type ClusterManager struct {
	Mut    sync.RWMutex
	ZkConn *zk.Conn
	// table->TableInfoWatcher
	Tables gcache.Cache
	// metaAddrs->metaManager
	Metas map[string]*session.MetaManager
}

// zkContext cancels the goroutine that's watching the zkNode, when the watcher
// is evicted by LRU.
type zkContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type TableInfoWatcher struct {
	tableName   string
	clusterName string
	metaAddrs   string
	event       <-chan zk.Event
	ctx         zkContext
}

func initClusterManager() {
	zkRequestCount = metrics.RegisterMeterWithTags("zk_request_count", []string{"table"})

	zkAddrs := config.GlobalConfig.ZookeeperOpts.Address
	zkConn, _, err := zk.Connect(config.GlobalConfig.ZookeeperOpts.Address,
		time.Duration(config.GlobalConfig.ZookeeperOpts.Timeout*1000000)) // the config value unit is ms, but zk request ns)
	if err != nil {
		logrus.Panicf("failed to connect to zookeeper \"%s\": %s", zkAddrs, err)
	}

	tables := gcache.New(config.GlobalConfig.ZookeeperOpts.WatcherCount).LRU().EvictedFunc(func(key interface{}, value interface{}) {
		value.(*TableInfoWatcher).ctx.cancel()
		logrus.Debugf("[%s] zk watcher is evicted", key.(string))
	}).Build() // TODO(jiashuo1) consider set expire time
	globalClusterManager = &ClusterManager{
		ZkConn: zkConn,
		Tables: tables,
		Metas:  make(map[string]*session.MetaManager),
	}
}

// return (metaAddr, metaManager, error)
func (m *ClusterManager) getMeta(table string) (string, *session.MetaManager, error) {
	var addrs string
	var meta *session.MetaManager

	tableInfo, err := m.Tables.Get(table)
	if err == nil {
		addrs = tableInfo.(*TableInfoWatcher).metaAddrs
		meta = m.Metas[tableInfo.(*TableInfoWatcher).metaAddrs]
		if meta != nil {
			return addrs, meta, nil
		}
	}

	//TODO(jiashuo) perf-counter
	m.Mut.Lock()
	defer m.Mut.Unlock()
	tableInfo, err = m.Tables.Get(table)
	if err != nil {
		tableInfo, err = m.newTableInfo(table)
		if err != nil {
			logrus.Errorf("[%s] failed to get cluster info: %s", table, err)
			return "", nil, err
		}
		err = m.Tables.Set(table, tableInfo)
		if err != nil {
			logrus.Errorf("[%s] failed to update local cache cluster info: %s", table, err)
			return "", nil, base.ERR_INVALID_DATA
		}
	}
	tableInfoW := tableInfo.(*TableInfoWatcher)
	addrs = tableInfoW.metaAddrs
	meta = m.Metas[addrs]
	if meta == nil {
		metaList, err := parseToMetaList(addrs)
		if err != nil {
			logrus.Errorf("[%s] cluster addr[%s] format is err: %s", table, addrs, err)
			return "", nil, base.ERR_INVALID_DATA
		}
		meta = session.NewMetaManager(metaList, session.NewNodeSession)
		m.Metas[addrs] = meta
	}

	return addrs, meta, nil
}

// get table cluster info and watch it based table name from zk
// The zookeeper path layout:
// /<RegionPathRoot>/<table> =>
//                         {
//                           "cluster_name" : "clusterName",
//                           "meta_addrs" : "metaAddr1,metaAddr2,metaAddr3"
//                         }
func (m *ClusterManager) newTableInfo(table string) (*TableInfoWatcher, error) {
	zkRequestCount.UpdateWithTags([]string{table})

	path := fmt.Sprintf("%s/%s", config.GlobalConfig.ZookeeperOpts.Root, table)
	value, _, watcherEvent, err := m.ZkConn.GetW(path)
	zkAddrs := config.GlobalConfig.ZookeeperOpts.Address
	if err != nil {
		if err == zk.ErrNoNode {
			logrus.Errorf("[%s] cluster info doesn't exist on zk[%s(%s)], err: %s", table, zkAddrs, path, err)
			return nil, base.ERR_OBJECT_NOT_FOUND
		}
		logrus.Errorf("[%s] failed to get cluster info from zk[%s(%s)]: %s", table, zkAddrs, path, err)
		return nil, base.ERR_ZOOKEEPER_OPERATION
	}

	type clusterInfoStruct struct {
		Name      string `json:"cluster_name"`
		MetaAddrs string `json:"meta_addrs"`
	}
	var cluster = &clusterInfoStruct{}
	err = json.Unmarshal(value, cluster)
	if err != nil {
		logrus.Errorf("[%s] cluster info on zk[%s(%s)] format is invalid, err = %s", table, zkAddrs, path, err)
		return nil, base.ERR_INVALID_DATA
	}

	ctx, cancel := context.WithCancel(context.Background())
	tableInfo := &TableInfoWatcher{
		tableName:   table,
		clusterName: cluster.Name,
		metaAddrs:   cluster.MetaAddrs,
		event:       watcherEvent,
		ctx: zkContext{
			ctx:    ctx,
			cancel: cancel,
		},
	}
	go m.watchTableInfoChanged(tableInfo)

	return tableInfo, nil
}

func (m *ClusterManager) watchTableInfoChanged(watcher *TableInfoWatcher) {
	select {
	case event := <-watcher.event:
		tableName, err := parseToTableName(event.Path)
		if err != nil {
			logrus.Panicf("zk path \"%s\" is corrupt, unable to parse table name: %s", event.Path, err)
		}
		if event.Type == zk.EventNodeDataChanged {
			tableInfo, err := m.newTableInfo(tableName)
			if err != nil {
				logrus.Panicf("[%s] failed to get cluster info when trigger watcher: %s", tableName, err)
			}
			m.Mut.Lock()
			err = m.Tables.Set(tableName, tableInfo)
			m.Mut.Unlock()
			if err != nil {
				logrus.Panicf("[%s] failed to update local cache cluster info to %s(%s): %s",
					tableName, tableInfo.clusterName, tableInfo.metaAddrs, err)
			}
			logrus.Infof("[%s] local cache cluster info is updated to %s(%s)", tableName,
				tableInfo.clusterName, tableInfo.metaAddrs)
		} else if event.Type == zk.EventNodeDeleted {
			m.Mut.Lock()
			if m.Tables.Has(tableName) {
				success := m.Tables.Remove(tableName)
				if !success {
					logrus.Panicf("[%s] failed to remove local cache cluster info", tableName)
				}
			}
			m.Mut.Unlock()
			logrus.Infof("[%s] local cache cluster info is removed", tableName)
		} else {
			logrus.Errorf("[%s] unexpected zk event, type = %s.", tableName, event.Type.String())
		}

	case <-watcher.ctx.ctx.Done():
		return
	}
}

func parseToMetaList(metaAddrs string) ([]string, error) {
	result := strings.Split(metaAddrs, ",")
	if len(result) < 2 {
		return []string{}, fmt.Errorf("the meta addrs[%s] is invalid", metaAddrs)
	}
	return result, nil
}

// parseToTableName extracts table name from the zookeeper path.
// The zookeeper path layout:
// /<RegionPathRoot>
//            /<table1> => {JSON}
//            /<table2> => {JSON}
func parseToTableName(path string) (string, error) {
	result := strings.Split(path, "/")
	if len(result) < 3 {
		return "", fmt.Errorf("the path[%s] is invalid", path)
	}

	return result[len(result)-1], nil
}
