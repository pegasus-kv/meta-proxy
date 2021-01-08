package metrics

import (
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/sirupsen/logrus"
)

// Gauge is the generic type for performance counters which can be `add` or `delete`, for example, used for current
// connection number
type Gauge interface {
	Add(value int64)
	Inc()

	Decrease(value int64)
	Dec()
}

// Meter is the generic type for performance counters which only can be `add`, for example, used for request-rate(qps)
type Meter interface {
	Update()
}

// init metric base config
func Init() {
	mtype := config.GlobalConfig.MetricsOpts.Type
	if mtype == "prometheus" {
		go startPromHTTPServer()
		return
	} else if mtype == "falcon" {
		return
	}
	logrus.Panic("no support metric type")
}

func RegisterGauge(name string) Gauge {
	mtype := config.GlobalConfig.MetricsOpts.Type
	if mtype == "prometheus" {
		labelsName, labelsValue := parseTags()
		return registerPromGauge(name, labelsName, labelsValue)
	} else if mtype == "falcon" {
		return registerFalconGauge(name)
	}
	logrus.Panicf("no support metric type: %s", mtype)
	return nil
}

func RegisterMeter(name string) Meter {
	mtype := config.GlobalConfig.MetricsOpts.Type
	if mtype == "prometheus" {
		labelsName, labelsValue := parseTags()
		return registerPromMeter(name, labelsName, labelsValue)
	} else if mtype == "falcon" {
		return registerFalconMeter(name)
	}
	logrus.Panicf("no support metric type: %s", mtype)
	return nil
}

// parse map tags into prometheus labels format->(LabelsName, LabelsValue)
func parseTags() ([]string, []string) {
	var labelsName []string
	var labelsValue []string
	for key, value := range config.GlobalConfig.MetricsOpts.Tags {
		labelsName = append(labelsName, key)
		labelsValue = append(labelsValue, value)
	}
	return labelsName, labelsValue
}
