package collector

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.InitConfig("../meta-proxy.yml")
	config.Config.PerfCounterOpt.Type = "prometheus"
}

func TestParseTags(t *testing.T) {
	names, values := parseTags()
	assert.Contains(t, names, "region")
	assert.Contains(t, names, "service")
	assert.Contains(t, values, "c3tst_staging")
	assert.Contains(t, values, "meta_proxy")
}

func TestPerfCounter(t *testing.T) {
	InitPerfCounter()
	// mock the promGauge counter: TableWatcherEvictCounter = 0
	TableWatcherEvictCounter.(*promGauge).Add(100)
	TableWatcherEvictCounter.(*promGauge).Incr()
	TableWatcherEvictCounter.(*promGauge).Delete(100)
	TableWatcherEvictCounter.(*promGauge).Decrease()

	// mock the promMeter: counter = 1
	ClientQueryConfigQPS.(*promMeter).Update()
	time.Sleep(1000000000)
	resp, err := http.Get("http://localhost:1988/metrics")
	assert.Nil(t, err)
	// the resp page content like: "counter value \n counter value \n"
	body, _ := ioutil.ReadAll(resp.Body)
	result := strings.Split(string(body), "\n")
	assert.Contains(t, result, "table_watcher_cache_evict_count{region=\"c3tst_staging\",service=\"meta_proxy\"} 0")
	assert.Contains(t, result, "client_query_config_request_qps{region=\"c3tst_staging\",service=\"meta_proxy\"} 1")
}
