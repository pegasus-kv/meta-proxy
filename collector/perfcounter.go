package collector

import (
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/sirupsen/logrus"
)

var pfcType string

var (
	TableWatcherEvictCounter Counter
	ClientConnectionCounter  Counter
)

type Counter interface {
	Add(value int64)
	Incr()
}

func InitPerfCounter() {
	pfcType = config.Cfg.Pfc.Type
	TableWatcherEvictCounter = registerCounter("table_watcher_cache_evict_count").(Counter)
	ClientConnectionCounter = registerCounter("client_connection_count").(Counter)

	if pfcType == "prometheus" {
		go start()
		return
	} else if pfcType == "falcon" {
		return
	}
	logrus.Panic("no support monitor type")
}

func registerCounter(counterName string) interface{} {
	if pfcType == "prometheus" {
		labelsName, labelsValue := parseTags()
		return registerPromCounter(counterName, labelsName, labelsValue)
	} else if pfcType == "falcon" {
		return registerFalconCounter(counterName)
	}
	logrus.Panic("no support monitor type")
	return nil
}

// parse map tags into prometheus labels format->(LabelsName, LabelsValue)
func parseTags() ([]string, []string) {
	var labelsName []string
	var labelsValue []string
	for key, value := range config.Cfg.Pfc.Tags {
		labelsName = append(labelsName, key)
		labelsValue = append(labelsValue, value)
	}
	return labelsName, labelsValue
}
