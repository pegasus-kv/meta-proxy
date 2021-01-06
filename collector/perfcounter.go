package collector

import (
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/sirupsen/logrus"
)

var perfCounterType string

var (
	TableWatcherEvictCounter Gauge
	ClientConnectionCounter  Gauge
	ClientQueryConfigQPS     Meter
)

// Gauge is the generic type for performance counters which can be `add` or `delete`, for example, used for current
// connection number
type Gauge interface {
	Add(value int64)
	Incr()

	Delete(value int64)
	Decrease()
}

// Meter is the generic type for performance counters which only can be `add`, for example, used for request-rate(qps)
type Meter interface {
	Update()
}

// init perfcounter base config
func InitPerfCounter() {
	perfCounterType = config.Config.PerfCounterOpt.Type
	TableWatcherEvictCounter = registerGauge("table_watcher_cache_evict_count").(Gauge)
	ClientConnectionCounter = registerGauge("client_connection_count").(Gauge)
	ClientQueryConfigQPS = registerMeter("client_query_config_request_qps").(Meter)

	if perfCounterType == "prometheus" {
		go startPromHTTPServer()
		return
	} else if perfCounterType == "falcon" {
		return
	}
	logrus.Panic("no support monitor type")
}

func registerGauge(counterName string) interface{} {
	if perfCounterType == "prometheus" {
		labelsName, labelsValue := parseTags()
		return registerPromGauge(counterName, labelsName, labelsValue)
	} else if perfCounterType == "falcon" {
		return registerFalconMetric(counterName)
	}
	logrus.Panic("no support monitor type")
	return nil
}

func registerMeter(counterName string) interface{} {
	if perfCounterType == "prometheus" {
		labelsName, labelsValue := parseTags()
		return registerPromMeter(counterName, labelsName, labelsValue)
	} else if perfCounterType == "falcon" {
		return registerFalconMetric(counterName)
	}
	logrus.Panic("no support monitor type")
	return nil
}

// parse map tags into prometheus labels format->(LabelsName, LabelsValue)
func parseTags() ([]string, []string) {
	var labelsName []string
	var labelsValue []string
	for key, value := range config.Config.PerfCounterOpt.Tags {
		labelsName = append(labelsName, key)
		labelsValue = append(labelsValue, value)
	}
	return labelsName, labelsValue
}
