package collector

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	TableWatcherEvictCounter = registerCounter(
		"table_watcher_evict_count",
		"help",
		[]string{"service"},
		[]string{"pegasus_meta_proxy"})
)

var metrics []*Metric

func (p *Metric) Incr()  {
	p.MetricWithLabel.(prometheus.Counter).Inc()
}

func (p *Metric) delete() {
	switch m := p.Metric.(type) {
	case *prometheus.CounterVec:
		m.DeleteLabelValues(p.LabelValue...)
	case *prometheus.HistogramVec:
		m.DeleteLabelValues(p.LabelValue...)
	default:
		logrus.Panicf("not support metric type")
	}
}

func Start() {
	defer cleanMetricLabel()

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}

func registerHistogramMetric(CounterName string, CounterHelp string, labelName []string, labelValue []string) *Metric {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    CounterName,
		Help:    CounterHelp,
		Buckets: prometheus.DefBuckets,
	}, labelName)
	histogramWithLabel := histogram.WithLabelValues(labelValue...)
	prometheus.MustRegister(histogram)

	promMetric := &Metric{
		LabelName:       labelName,
		LabelValue:      labelValue,
		Metric:          histogram,
		MetricWithLabel: histogramWithLabel,
	}
	metrics = append(metrics, promMetric)
	return promMetric
}

func cleanMetricLabel() {
	for _, m := range metrics {
		m.delete()
	}
}
