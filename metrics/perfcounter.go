package metrics

import (
	"strings"

	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/sirupsen/logrus"
)

// Gauge is the generic type for performance counters which can be `add` or `delete`, for example, used for current
// connection number
type Gauge interface {
	Add(value int64)
	AddWithTags(tagsValue []string, counterValue int64)

	Inc()
	IncWithTags(tagsValue []string)

	Sub(value int64)
	SubWithTags(tagsValue []string, counterValue int64)

	Dec()
	DecWithTags(tagsValue []string)
}

// Meter is the generic type for performance counters which only can be `add`, for example, used for request-rate(qps)
type Meter interface {
	Update()
	UpdateWithTags(tagsValue []string)
}

// Init metric base config
func Init() {
	mtype := config.GlobalConfig.MetricsOpts.Type
	if mtype == "prometheus" {
		go startPromHTTPServer()
		return
	} else if mtype == "falcon" {
		return
	}
	logrus.Panicf("no support tags type: %s", mtype)
}

// RegisterGauge using counter name with default tags in config
func RegisterGauge(counterName string) Gauge {
	return RegisterGaugeWithTags(counterName, []string{})
}

// RegisterGaugeWithTags using counter name with custom tags and default tags of config
func RegisterGaugeWithTags(counterName string, tagsName []string) Gauge {
	mtype := config.GlobalConfig.MetricsOpts.Type
	if mtype == "prometheus" {
		return registerPromGauge(counterName, tagsName)
	} else if mtype == "falcon" {
		return registerFalconGauge(counterName, tagsName)
	}
	logrus.Panicf("no support tagsName type: %s", mtype)
	return nil
}

// RegisterMeter using counter name with default tags in config
func RegisterMeter(counterName string) Meter {
	return RegisterMeterWithTags(counterName, []string{})
}

// RegisterMeterWithTags using counter name with custom tags and default tags of config
func RegisterMeterWithTags(counterName string, tagsName []string) Meter {
	mtype := config.GlobalConfig.MetricsOpts.Type
	if mtype == "prometheus" {
		return registerPromMeter(counterName, tagsName)
	} else if mtype == "falcon" {
		return registerFalconMeter(counterName, tagsName)
	}
	logrus.Panicf("no support tags type: %s", mtype)
	return nil
}

// join default tagsName of config and custom tagsName
func combineConfigTagsName(tagsName []string) []string {
	for _, tag := range config.GlobalConfig.MetricsOpts.Tags {
		tagsName = append(tagsName, strings.Split(tag, "=")[0])
	}
	return tagsName
}

// join default tagsValue of config and custom tagsValue
func combineConfigTagsValue(tagsValue []string) []string {
	for _, tag := range config.GlobalConfig.MetricsOpts.Tags {
		tagsValue = append(tagsValue, strings.Split(tag, "=")[1])
	}
	return tagsValue
}
