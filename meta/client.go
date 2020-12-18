package meta

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/pegalog"
	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/bluele/gcache"
	"github.com/samuel/go-zookeeper/zk"
)

var zkAddrs = []string{""}
var zkTimeOut = 10
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
				tableName := event.Path // TODO(jiashuo1) split to get table name
				addr := clusterManager.getClusterAddr(tableName)
				clusterManager.Lock.Lock()
				err := clusterManager.Tables.Set(tableName, addr)
				if err != nil {
					panic("TODO")
				}
				clusterManager.Lock.Unlock()
			} else {
				// TODO(jiashuo1)
			}
		}()
	})

	zkConn, _, err := zk.Connect(zkAddrs, time.Duration(zkTimeOut), option)
	if err != nil {
		panic("")
	}
	tables := gcache.New(1028).LFU().Build() // TODO(jiashuo1) can set expire time
	clusterManager = &ClusterManager{
		ZkConn: zkConn,
		Tables: tables,
		Metas:  make(map[string]*session.MetaManager),
	}
}

func (m *ClusterManager) getMetaConnector(table string) *session.MetaManager {
	metaAddrs, err := clusterManager.Tables.Get(table)
	if err != nil {
		panic("TODO")
	}
	meta := clusterManager.Metas[metaAddrs.(string)]
	if meta == nil {
		clusterManager.Lock.Lock()

		if len(metaAddrs.(string)) == 0 {
			metaAddrs = m.getClusterAddr(table)
			err = clusterManager.Tables.Set(table, metaAddrs)
			if err != nil {
				panic("TODO")
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
		panic("TODO(jiashuo)")
	}

	type tableInfoStruct struct {
		ClusterName string
		MetaAddrs   string
	}

	var tableInfo = &tableInfoStruct{}
	err = json.Unmarshal(value, tableInfo)
	if err != nil {
		panic("TODO(jiashuo)")
	}

	return tableInfo.MetaAddrs
}

func str2slice(meta string) []string {
	result := strings.Split(meta, ",")
	if len(result) == 0 {
		pegalog.GetLogger().Fatal(fmt.Sprintf("Invalid meta address %s", meta))
	}
	return result
}
