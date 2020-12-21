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

var zkAddrs = []string{""}
var zkTimeOut = 100000000000000
var zkRoot = "/pegasus-cluster"

var clusterManager *ClusterManager

type ClusterManager struct {
	Lock   sync.Mutex
	ZkConn *zk.Conn
	// table->addrs
	Tables gcache.Cache
	// addrs->metaManager
	Metas map[string]*session.MetaManager
}

func initClusterManager() {
	option := zk.WithEventCallback(func(event zk.Event) {
		go func() {
			if event.Type == zk.EventNodeDataChanged {
				tableName := getTableName(event.Path) // TODO(jiashuo1) split to get table name
				addr := clusterManager.getClusterAddr(tableName)
				clusterManager.Lock.Lock()
				err := clusterManager.Tables.Set(tableName, addr)
				clusterManager.Lock.Unlock()
				if err != nil {
					panic("TODO")
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

	tables := gcache.New(1028).LFU().Build() // TODO(jiashuo1) can set expire time
	clusterManager = &ClusterManager{
		ZkConn: zkConn,
		Tables: tables,
		Metas:  make(map[string]*session.MetaManager),
	}
}

func (m *ClusterManager) getMetaConnector(table string) *session.MetaManager {
	var meta *session.MetaManager

	metaAddrs, err := clusterManager.Tables.Get(table)
	if err != nil {
		// TODO(jiashuo) log
		clusterManager.Lock.Lock()
		metaAddrs, err = clusterManager.Tables.Get(table)
		if err != nil {
			metaAddrs = m.getClusterAddr(table)
			err := clusterManager.Tables.Set(table, metaAddrs.(string))
			if err != nil {
				clusterManager.Lock.Unlock()
				panic(err)
			}
		}

		meta = clusterManager.Metas[metaAddrs.(string)]
		if meta == nil {
			meta = session.NewMetaManager(str2slice(metaAddrs.(string)), session.NewNodeSession)
			clusterManager.Metas[metaAddrs.(string)] = meta
		}

		clusterManager.Lock.Unlock()
	}
	return meta
}

// TODO(jiashuo) get cluster addr based table name
func (m *ClusterManager) getClusterAddr(table string) string {
	path := fmt.Sprintf("%s/%s", zkRoot, table)
	value, _, _, err := clusterManager.ZkConn.GetW(path)
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

	return tableInfo.MetaAddrs
}

func str2slice(meta string) []string {
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
