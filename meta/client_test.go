package meta

import (
	"github.com/samuel/go-zookeeper/zk"
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
	ret, _, _, _ := clusterManager.ZkConn.ExistsW(zkRoot)
	if !ret {
		_, err := clusterManager.ZkConn.Create(zkRoot, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			panic(err)
		}
	}

	for _, test := range tests {
		ret, stat, _, _ := clusterManager.ZkConn.ExistsW(test.path)
		if ret {
			_ = clusterManager.ZkConn.Delete(test.path, stat.Version)
		}
		_, err := clusterManager.ZkConn.Create(test.path, []byte(test.data), 0, acls)
		if err != nil {
			panic(err)
		}
	}
}

func TestZookeeper(t *testing.T) {
	for _, test := range tests {
		addrs, _ := clusterManager.getTableInfo(test.table)
		assert.Equal(t, test.addr, addrs.metaAddrs)

		_, _ = clusterManager.getMetaConnector(test.table)
		cacheWatcher, _ := clusterManager.Tables.Get(test.table)
		assert.Equal(t, test.addr, cacheWatcher.(*TableInfoWithWatcher).metaAddrs)

		for _, update := range updates {
			// update zookeeper node data and trigger the watch event update local cache
			_, stat, _ := clusterManager.ZkConn.Get(test.path)
			_, err := clusterManager.ZkConn.Set(test.path, []byte(update.data), stat.Version)
			if err != nil {
				panic(err)
			}

			// local cache will change to new meta addr
			time.Sleep(time.Duration(10000000))
			cacheWatcher, _ = clusterManager.Tables.Get(test.table)
			assert.Equal(t, update.addr, cacheWatcher.(*TableInfoWithWatcher).metaAddrs)
		}
	}

	assert.Equal(t, clusterManager.Tables.Len(true), 2)
}
