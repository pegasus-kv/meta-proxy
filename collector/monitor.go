package collector

import "github.com/sirupsen/logrus"

// TODO(jiashuo1) store config
var monitorType string
var monitorTag map[string]string

type Counter interface {
	Incr()
}

func init() {
	if monitorType == "prometheus" {
		start()
	} else if monitorType == "falcon" {
		return
	}
	logrus.Panic("no support monitor type")
}

func registerCounter(counterName string) interface{} {
	if monitorType == "prometheus" {
		// TODO(jiashuo1) tag should be same with falcon and get tag from config
		return registerPromCounter(counterName, "counterHelp",
			[]string{"service"}, []string{"meta-proxy"})
	} else if monitorType == "falcon" {
		return registerFalconCounter(counterName)
	}
	logrus.Panic("no support monitor type")
	return nil
}
