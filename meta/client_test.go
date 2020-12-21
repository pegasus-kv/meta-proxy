package meta

import (
	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testZkaddrs = []string{"127.0.0.1:22181"}
var testTable = "temp"
var tablePath = zkRoot + "/" + testTable
var testData = "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}"

func TestInit(t *testing.T) {
	zkAddrs = testZkaddrs
	initClusterManager()

	var data = []byte(testData)
	acls := zk.WorldACL(zk.PermAll)

	ret, _, _, _ := clusterManager.ZkConn.ExistsW(zkRoot)
	if !ret {
		_, err := clusterManager.ZkConn.Create(zkRoot, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			panic(err)
		}
	}

	ret, stat, _, _ := clusterManager.ZkConn.ExistsW(tablePath)
	if ret {
		_ = clusterManager.ZkConn.Delete(tablePath, stat.Version)
	}
	_, err := clusterManager.ZkConn.Create(tablePath, data, 0, acls)
	if err != nil {
		panic(err)
	}
}

func TestZookeeper(t *testing.T) {
	addrs := clusterManager.getClusterAddr(testTable)
	assert.Equal(t, "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603", addrs)

	clusterManager.getMetaConnector(testTable)
	cacheAddrs, _ := clusterManager.Tables.Get("temp")
	assert.Equal(t, "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603", cacheAddrs)

	// update zookeeper node data and trigger the watch event update local cache
	newData := []byte("{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.1.1.1:34601,127.1.1.1:34602,127.1.1.1:34603\"}")
	_, err := clusterManager.ZkConn.Set(tablePath, newData, 0)
	if err != nil {
		panic(err)
	}

	// local cache will change the new meta addr
	time.Sleep(time.Duration(10000000))
	cacheAddrs, _ = clusterManager.Tables.Get("temp")
	assert.Equal(t, "127.1.1.1:34601,127.1.1.1:34602,127.1.1.1:34603", cacheAddrs)
}
