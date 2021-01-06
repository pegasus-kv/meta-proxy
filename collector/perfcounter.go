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

type Gauge interface {
	Add(value int64)
	Incr()

	Delete(value int64)
	Decrease()
}

type Meter interface {
	Update()
}

func InitPerfCounter() {
	perfCounterType = config.Config.PerfCounterOpt.Type
	TableWatcherEvictCounter = registerCounter("table_watcher_cache_evict_count").(Gauge)
	ClientConnectionCounter = registerCounter("client_connection_count").(Gauge)
	ClientQueryConfigQPS = registerMeter("client_query_config_request_qps").(Meter)

	if perfCounterType == "prometheus" {
		go startPromHTTPServer()
		return
	} else if perfCounterType == "falcon" {
		return
	}
	logrus.Panic("no support monitor type")
}

// report the current total count
func registerCounter(counterName string) interface{} {
	if perfCounterType == "prometheus" {
		labelsName, labelsValue := parseTags()
		return registerPromGauge(counterName, labelsName, labelsValue)
	} else if perfCounterType == "falcon" {
		return registerFalconMetric(counterName)
	}
	logrus.Panic("no support monitor type")
	return nil
}

// report the current request rate
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
