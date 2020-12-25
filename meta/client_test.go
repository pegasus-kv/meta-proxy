package meta

import (
	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testZkAddrs = []string{"127.0.0.1:22181"}

type testCase struct {
	table string
	addr  string
	path  string
	data  string
}

var tests = []testCase{
	{
		table: "temp",
		addr:  "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		path:  zkRoot + "/temp",
		data:  "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}"},
	{
		table: "stat",
		addr:  "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		path:  zkRoot + "/stat",
		data:  "{\"cluster_name\": \"onebox\", \"meta_addrs\": \"127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603\"}"},
	{
		table: "test",
		addr:  "127.0.0.1:34601,127.0.0.1:34602,127.0.0.1:34603",
		path:  zkRoot + "/test",
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

func init() {
	zkAddrs = testZkAddrs
	zkWatcherCount = 2
	initClusterManager()

	acls := zk.WorldACL(zk.PermAll)
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

func TestZookeeper(t *testing.T) {
	for _, test := range tests {
		addrs, _ := globalClusterManager.getTableInfo(test.table)
		assert.Equal(t, test.addr, addrs.metaAddrs)

		_, _ = globalClusterManager.getMetaConnector(test.table)
		cacheWatcher, _ := globalClusterManager.Tables.Get(test.table)
		assert.Equal(t, test.addr, cacheWatcher.(*TableInfoWatcher).metaAddrs)

		for _, update := range updates {
			// update zookeeper node data and trigger the watch event update local cache
			_, stat, _ := globalClusterManager.ZkConn.Get(test.path)
			_, err := globalClusterManager.ZkConn.Set(test.path, []byte(update.data), stat.Version)
			if err != nil {
				panic(err)
			}

			// local cache will change to new meta addr
			time.Sleep(time.Duration(10000000))
			cacheWatcher, _ = globalClusterManager.Tables.Get(test.table)
			assert.Equal(t, update.addr, cacheWatcher.(*TableInfoWatcher).metaAddrs)
		}
	}

	assert.Equal(t, globalClusterManager.Tables.Len(true), 2)
}
