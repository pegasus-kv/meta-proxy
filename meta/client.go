package meta

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/bluele/gcache"
	"github.com/samuel/go-zookeeper/zk"
)

// TODO(jiashuo) store config file
var zkAddrs = []string{""}
var zkTimeOut = 1000000000 // unit ns, equal 1s
var zkRoot = "/pegasus-cluster"
var zkWatcherCount = 1024

var clusterManager *ClusterManager

type ClusterManager struct {
	Lock   sync.Mutex
	ZkConn *zk.Conn
	// table->TableInfoWithWatcher
	Tables gcache.Cache
	// metaAddrs->metaManager
	Metas map[string]*session.MetaManager
}

type TableInfoWithWatcher struct {
	clusterName string
	metaAddrs   string
	event       <-chan zk.Event
}

// TODO(jishuo1) may need change log module
func initClusterManager() {
	option := zk.WithEventCallback(func(event zk.Event) {
		go func() {
			tableName, err := getTableName(event.Path)
			if err != nil {
				return
			}
			if event.Type == zk.EventNodeDataChanged {
				tableInfo, err := clusterManager.getTableInfo(tableName)
				if err != nil {
					log.Printf("[%s] get cluster info failed when triger watcher: %s", tableName, err)
					return
				}
				log.Printf("[%s] cluster info is updated to %s(%s)", tableName, tableInfo.clusterName, tableInfo.metaAddrs)
				clusterManager.Lock.Lock()
				err = clusterManager.Tables.Set(tableName, tableInfo)
				clusterManager.Lock.Unlock()
				if err != nil {
					log.Printf("[%s] cluster info local cache updated to %s(%s) failed: %s", tableName, tableInfo.clusterName, tableInfo.metaAddrs, err)
				}
			} else if event.Type == zk.EventNodeDeleted {
				log.Printf("[%s] cluster info is removed.", tableName)
				clusterManager.Lock.Lock()
				success := clusterManager.Tables.Remove(tableName)
				clusterManager.Lock.Unlock()
				if !success {
					log.Printf("[%s] cluster info local cache removed failed!", tableName)
				}
			} else {
				log.Printf("[%s] cluster info is updated, type = %s.", tableName, event.Type.String())
			}
		}()
	})

	zkConn, _, err := zk.Connect(zkAddrs, time.Duration(zkTimeOut), option)
	if err != nil {
		panic(fmt.Errorf("connect to %s failed: %s", zkAddrs, err))
	}

	tables := gcache.New(zkWatcherCount).LRU().Build() // TODO(jiashuo1) consider set expire time
	clusterManager = &ClusterManager{
		ZkConn: zkConn,
		Tables: tables,
		Metas:  make(map[string]*session.MetaManager),
	}
}

func (m *ClusterManager) getMetaConnector(table string) (*session.MetaManager, error) {
	var metaConnector *session.MetaManager

	tableInfo, err := clusterManager.Tables.Get(table)
	if err == nil {
		metaConnector = clusterManager.Metas[tableInfo.(*TableInfoWithWatcher).metaAddrs]
		if metaConnector != nil {
			return metaConnector, nil
		}
	}

	log.Printf("[%s] can't get cluster info from local cache, try fetch from zk.", table)
	clusterManager.Lock.Lock()
	tableInfo, err = clusterManager.Tables.Get(table)
	if err != nil {
		tableInfo, err = m.getTableInfo(table)
		if err != nil {
			return nil, err
		}
		err = clusterManager.Tables.Set(table, tableInfo)
		if err != nil {
			log.Printf("[%s] cluster info update local cache failed: %s", table, err)
		}
	}

	metaAddrs := tableInfo.(*TableInfoWithWatcher).metaAddrs
	metaConnector = clusterManager.Metas[metaAddrs]
	if metaConnector == nil {
		metaList, err := getMetaList(metaAddrs)
		if err != nil {
			return nil, err
		}
		metaConnector = session.NewMetaManager(metaList, session.NewNodeSession)
		clusterManager.Metas[metaAddrs] = metaConnector
	}

	clusterManager.Lock.Unlock()
	return metaConnector, nil
}

// get table cluster info and watch it based table name from zk
func (m *ClusterManager) getTableInfo(table string) (*TableInfoWithWatcher, error) {
	path := fmt.Sprintf("%s/%s", zkRoot, table)
	value, _, watcherEvent, err := clusterManager.ZkConn.GetW(path)
	if err != nil {
		return nil, fmt.Errorf("get table info from %s failed: %s", zkAddrs, err)
	}

	type clusterInfoStruct struct {
		Name      string `json:"cluster_name"`
		MetaAddrs string `json:"meta_addrs"`
	}
	var cluster = &clusterInfoStruct{}
	err = json.Unmarshal(value, cluster)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json failed: %s", err)
	}

	tableInfo := TableInfoWithWatcher{
		clusterName: cluster.Name,
		metaAddrs:   cluster.MetaAddrs,
		event:       watcherEvent,
	}

	return &tableInfo, nil
}

func getMetaList(metaAddrs string) ([]string, error) {
	result := strings.Split(metaAddrs, ",")
	if len(result) < 2 {
		return []string{}, fmt.Errorf("the meta addrs[%s] is invalid", metaAddrs)
	}
	return result, nil
}

func getTableName(path string) (string, error) {
	result := strings.Split(path, "/")
	if len(result) < 2 {
		return "", fmt.Errorf("the path[%s] is invalid", path)
	}

	return result[len(result)-1], nil
}
