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
}

func TestParseTags(t *testing.T) {
	names, values := parseTags()
	assert.Contains(t, names, "region")
	assert.Contains(t, names, "service")
	assert.Contains(t, values, "c3tst_staging")
	assert.Contains(t, values, "meta_proxy")
}

func TestPrometheus(t *testing.T) {
	config.Config.PerfCounterOpt.Type = "prometheus"
	InitPerfCounter()
	// mock the promGauge counter: TableWatcherEvictCounter = 0
	TableWatcherEvictCounter.Add(100)
	TableWatcherEvictCounter.Incr()
	TableWatcherEvictCounter.Delete(100)
	TableWatcherEvictCounter.Decrease()

	// mock the promMeter: counter = 1
	ClientQueryConfigQPS.(*promMeter).Update()
	time.Sleep(1000000000)
	resp, err := http.Get("http://localhost:9091/metrics")
	assert.Nil(t, err)
	// the resp page content like: "counter value \n counter value \n"
	body, _ := ioutil.ReadAll(resp.Body)
	result := strings.Split(string(body), "\n")
	assert.Contains(t, result, "table_watcher_cache_evict_count{region=\"c3tst_staging\",service=\"meta_proxy\"} 0")
	assert.Contains(t, result, "client_query_config_request_qps{region=\"c3tst_staging\",service=\"meta_proxy\"} 1")
}

func TestFalcon(t *testing.T) {
	config.Config.PerfCounterOpt.Type = "falcon"
	InitPerfCounter()
	// mock the falconGauge counter: TableWatcherEvictCounter = 0
	TableWatcherEvictCounter.(*FalconMetric).Add(100)
	TableWatcherEvictCounter.(*FalconMetric).Incr()
	TableWatcherEvictCounter.(*FalconMetric).Delete(100)
	TableWatcherEvictCounter.(*FalconMetric).Decrease()

	// mock the falconMeter: counter = 1
	ClientQueryConfigQPS.(*FalconMetric).Update()
}
