package collector

import (
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/sirupsen/logrus"
)

var pfcType string

var (
	TableWatcherEvictCounter = registerCounter(
		"table_watcher_cache_evict_count").(Counter)
)

type Counter interface {
	Incr()
}

func InitPerfCounter() {
	pfcType = config.Cfg.Pfc.Type
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
		labelsName = append(labelsValue, value)
	}
	return labelsName, labelsValue
}
