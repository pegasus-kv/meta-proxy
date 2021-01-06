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

func TestPerfCounter(t *testing.T) {
	InitPerfCounter()
	TableWatcherEvictCounter.(*PromCounter).Incr()
	time.Sleep(1000000000)
	resp, err := http.Get("http://localhost:8080/metrics")
	assert.Nil(t, err)
	// the resp page content like: "counter value \n counter value \n"
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t,
		strings.Split(string(body), "\n"),
		"table_watcher_cache_evict_count{region=\"c3tst_staging\",service=\"meta_proxy\"} 1")
}
