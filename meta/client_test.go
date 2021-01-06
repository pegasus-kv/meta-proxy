package meta

import (
	"fmt"
	"github.com/pegasus-kv/meta-proxy/collector"
	"github.com/pegasus-kv/meta-proxy/config"
	"io"
	"os"
	"testing"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/go-zookeeper/zk"
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
	config.InitConfig("../meta-proxy.yml")
	collector.InitPerfCounter()
	config.Cfg.Zk.WatcherCount = 2
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

func TestGetTable(t *testing.T) {
	// pass zkAddr can't be connected
	config.Cfg.Zk.Address = []string{"128.0.0.1:22171"}
	initClusterManager()
	_, err := globalClusterManager.newTableInfo("notExist")
	assert.Equal(t, err, base.ERR_ZOOKEEPER_OPERATION)

	config.Cfg.Zk.Address = []string{"127.0.0.1:22181"}
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
	config.Cfg.Zk.Address = []string{"127.0.0.1:22181"}
	initClusterManager()

	// first get connector which will init the cache and only store `stat` and `test` table watcher
	for _, test := range tests {
		_, _ = globalClusterManager.getMeta(test.table)
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
			meta, _ := globalClusterManager.getMeta(test.table)
			assert.NotNil(t, meta)
		}
	}
}

func TestZookeeperUpdate(t *testing.T) {
	for _, test := range tests {
		_, _ = globalClusterManager.getMeta(test.table)
		// update zookeeper node data and trigger the watch event update local cache
		for _, update := range updates {
			_, stat, _ := globalClusterManager.ZkConn.Get(test.path)
			_, err := globalClusterManager.ZkConn.Set(test.path, []byte(update.data), stat.Version)
			if err != nil {
				panic(err)
			}

			// local cache will change to new meta addr because the zk watcher
			time.Sleep(time.Duration(10000000))
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
	type table struct {
		path   string
		result error
	}

	tablePaths := []table{
		{
			path:   zkRoot + "/table",
			result: nil,
		},
		{
			path:   zkRoot,
			result: fmt.Errorf("the path[%s] is invalid", zkRoot),
		},
		{
			path:   zkRoot + "//table",
			result: fmt.Errorf("the path[%s] is invalid", zkRoot+"//table"),
		},
		{
			path:   zkRoot + "/table/name",
			result: fmt.Errorf("the path[%s] is invalid", zkRoot+"/table/name"),
		},
	}

	for _, tb := range tablePaths {
		ret, err := parseToTableName(tb.path)
		assert.Equal(t, err, tb.result)
		if err == nil {
			assert.Equal(t, "table", ret)
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
