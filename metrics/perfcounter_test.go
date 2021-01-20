package metrics

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
	config.Init("../config/yaml/local-meta-proxy.yml")
}

func TestParseTags(t *testing.T) {
	tagsName := combineConfigTagsName([]string{"table"})
	tagsValue := combineConfigTagsValue([]string{"temp"})
	assert.Equal(t, tagsName, []string{"table", "region", "service"})
	assert.Equal(t, tagsValue, []string{"temp", "local_tst", "meta_proxy"})
}

func TestPrometheus(t *testing.T) {
	config.GlobalConfig.MetricsOpts.Type = "prometheus"
	gaugeCounterWithTags := RegisterGaugeWithTags("promGaugeWithTagsTest", []string{"table"})
	meterCounterWithTags := RegisterMeterWithTags("promMeterWithTagsTest", []string{"table"})

	gaugeCounterNoTags := RegisterGauge("promGaugeNoTagsTest")
	meterCounterNoTags := RegisterMeter("promMeterNoTagsTest")
	Init()
	// mock the promGauge counter: gaugeCounterWithTags = 0
	gaugeCounterWithTags.AddWithTags([]string{"temp"}, 100)
	gaugeCounterWithTags.IncWithTags([]string{"temp"})
	gaugeCounterWithTags.SubWithTags([]string{"temp"}, 100)
	gaugeCounterWithTags.DecWithTags([]string{"temp"})

	gaugeCounterNoTags.Add(100)
	gaugeCounterNoTags.Inc()
	gaugeCounterNoTags.Sub(100)
	gaugeCounterNoTags.Dec()

	// mock the promMeter: meterCounter = 1
	meterCounterWithTags.UpdateWithTags([]string{"temp"})
	meterCounterNoTags.Update()
	time.Sleep(10000000)
	resp, err := http.Get("http://localhost:9091/metrics")
	assert.Nil(t, err)
	defer resp.Body.Close()
	// the resp page content like: "counter value \n counter value \n"
	body, _ := ioutil.ReadAll(resp.Body)
	result := strings.Split(string(body), "\n")
	assert.Contains(t, result, "promGaugeWithTagsTest{region=\"local_tst\",service=\"meta_proxy\",table=\"temp\"} 0")
	assert.Contains(t, result, "promMeterWithTagsTest{region=\"local_tst\",service=\"meta_proxy\",table=\"temp\"} 1")
	assert.Contains(t, result, "promGaugeNoTagsTest{region=\"local_tst\",service=\"meta_proxy\"} 0")
	assert.Contains(t, result, "promMeterNoTagsTest{region=\"local_tst\",service=\"meta_proxy\"} 1")
}

func TestFalcon(t *testing.T) {
	config.GlobalConfig.MetricsOpts.Type = "falcon"
	counterName := parseToCounterName("counterName", []string{"table"}, []string{"temp"})
	assert.Equal(t, "counterName,table=temp,region=local_tst,service=meta_proxy", counterName)

	gaugeCounterWithTags := RegisterGaugeWithTags("falconGaugeTest", []string{"table"})
	meterCounterWithTags := RegisterMeterWithTags("falconMeterTest", []string{"table"})
	gaugeCounterNoTags := RegisterGauge("falconGaugeTest")
	meterCounterNoTags := RegisterMeter("falconMeterTest")
	Init()
	// mock the falconGauge counter
	gaugeCounterWithTags.AddWithTags([]string{"temp"}, 100)
	gaugeCounterWithTags.IncWithTags([]string{"temp"})
	gaugeCounterWithTags.SubWithTags([]string{"temp"}, 100)
	gaugeCounterWithTags.DecWithTags([]string{"temp"})

	gaugeCounterNoTags.Add(100)
	gaugeCounterNoTags.Inc()
	gaugeCounterNoTags.Sub(100)
	gaugeCounterNoTags.Dec()

	// mock the falconMeter
	meterCounterWithTags.UpdateWithTags([]string{"temp"})
	meterCounterNoTags.Update()
}
