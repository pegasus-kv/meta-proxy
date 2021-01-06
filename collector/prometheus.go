package collector

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Label struct {
	LabelName  []string
	LabelValue []string
}

type PromCounter struct {
	Label
	Metric *prometheus.CounterVec
}

type PromHistogram struct {
	Label
	Metric *prometheus.HistogramVec
}

func (p *PromCounter) Add(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(float64(value))
}

func (p *PromCounter) Incr() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

func registerPromCounter(CounterName string, labelName []string, labelValue []string) *PromCounter {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: CounterName,
	}, labelName)
	prometheus.MustRegister(counter)

	return &PromCounter{
		Label: Label{
			LabelName:  labelName,
			LabelValue: labelValue,
		},
		Metric: counter,
	}
}

func start() {
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
