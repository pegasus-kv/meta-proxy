package collector

import "github.com/prometheus/client_golang/prometheus"
import "git.n.xiaomi.com/falcon-sdk/goperfcounter"

type Metric struct {
	LabelName  []string
	LabelValue []string

	Metric          interface{}
	MetricWithLabel interface{}
}

type Monitor interface {
	incr()
}

func registerCounter(monitorType string, CounterName string, CounterHelp string, labelName []string, labelValue []string) *Metric {
	if monitorType == "prometheus" {
		return registerPromCounter(CounterName, CounterHelp, labelName, labelValue)
	} else if monitorType == "falcon" {

	}
}

func registerPromCounter(CounterName string, CounterHelp string, labelName []string, labelValue []string) *Metric{
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: CounterName,
		Help: CounterHelp,
	}, labelName)
	counterWithLabel := counter.WithLabelValues(labelValue...)
	prometheus.MustRegister(counter)

	promMetric := &Metric{
		LabelName:       labelName,
		LabelValue:      labelValue,
		Metric:          counter,
		MetricWithLabel: counterWithLabel,
	}
	metrics = append(metrics, promMetric)
	return promMetric
}