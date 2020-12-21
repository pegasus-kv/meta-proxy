package meta

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/bluele/gcache"
	"github.com/samuel/go-zookeeper/zk"
)

// TODO(jiashuo) store config file
var zkAddrs = []string{""}
var zkTimeOut = 1000000000
var zkRoot = "/pegasus-cluster"
var zkWatcherCount = 1024

var clusterManager *ClusterManager

type ClusterManager struct {
	Lock   sync.Mutex
	ZkConn *zk.Conn
	// table->Watcher
	Tables gcache.Cache
	// addrs->metaManager
	Metas map[string]*session.MetaManager
}

type Watcher struct {
	addrs string
	event <-chan zk.Event
}

func initClusterManager() {
	option := zk.WithEventCallback(func(event zk.Event) {
		go func() {
			if event.Type == zk.EventNodeDataChanged {
				tableName := getTableName(event.Path)
				metaAddrs, event := clusterManager.getClusterAddr(tableName)
				clusterManager.Lock.Lock()
				err := clusterManager.Tables.Set(tableName, Watcher{
					addrs: metaAddrs,
					event: event,
				})
				clusterManager.Lock.Unlock()
				if err != nil {
					panic(err) // TODO(jiashuo) all the other panic will change to log
				}
			} else {
				// TODO(jiashuo1)
			}
		}()
	})

	zkConn, _, err := zk.Connect(zkAddrs, time.Duration(zkTimeOut), option)
	if err != nil {
		panic(err)
	}

	tables := gcache.New(zkWatcherCount).LFU().Build() // TODO(jiashuo1) can set expire time
	clusterManager = &ClusterManager{
		ZkConn: zkConn,
		Tables: tables,
		Metas:  make(map[string]*session.MetaManager),
	}
}

func (m *ClusterManager) getMetaConnector(table string) *session.MetaManager {
	var metaConnector *session.MetaManager

	watcher, err := clusterManager.Tables.Get(table)
	if err != nil {
		// TODO(jiashuo) log
		clusterManager.Lock.Lock()
		watcher, err = clusterManager.Tables.Get(table)
		if err != nil {
			metaAddrs, event := m.getClusterAddr(table)
			watcher = Watcher{
				addrs: metaAddrs,
				event: event,
			}
			err := clusterManager.Tables.Set(table, watcher)
			if err != nil {
				clusterManager.Lock.Unlock()
				panic(err)
			}
		}

		metaAddrs := watcher.(Watcher).addrs
		metaConnector = clusterManager.Metas[metaAddrs]
		if metaConnector == nil {
			metaConnector = session.NewMetaManager(getMetaAddrs(metaAddrs), session.NewNodeSession)
			clusterManager.Metas[metaAddrs] = metaConnector
		}

		clusterManager.Lock.Unlock()
	}
	return metaConnector
}

// TODO(jiashuo) get cluster addr based table name
func (m *ClusterManager) getClusterAddr(table string) (string, <-chan zk.Event) {
	path := fmt.Sprintf("%s/%s", zkRoot, table)
	value, _, watcherEvent, err := clusterManager.ZkConn.GetW(path)
	if err != nil {
		panic(err) // TODO(jiashuo) log
	}

	type tableInfoStruct struct {
		ClusterName string `json:"cluster_name"`
		MetaAddrs   string `json:"meta_addrs"`
	}

	var tableInfo = &tableInfoStruct{}
	err = json.Unmarshal(value, tableInfo)
	if err != nil {
		panic(err)
	}

	return tableInfo.MetaAddrs, watcherEvent
}

func getMetaAddrs(meta string) []string {
	result := strings.Split(meta, ",")
	if len(result) < 2 {
		// TODO(jiashuo) pegalog.GetLogger().Fatal(fmt.Sprintf("Invalid meta address %s", meta))
	}
	return result
}

func getTableName(path string) string {
	result := strings.Split(path, "/")
	if len(result) < 2 {
		panic("")
	}

	return result[len(result)-1]
}
