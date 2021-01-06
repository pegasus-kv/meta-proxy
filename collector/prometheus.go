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

/*** promGauge metric for reporting the current total number which can be "add" or "delete" ***/
type promGauge struct {
	Label
	Metric *prometheus.GaugeVec
}

func (p *promGauge) Add(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(float64(value))
}

func (p *promGauge) Incr() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

func (p *promGauge) Delete(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(-float64(value))
}

func (p *promGauge) Decrease() {
	p.Metric.WithLabelValues(p.LabelValue...).Dec()
}

/*** promMeter metric for reporting the rate number which only can be "add"
  and the rate get by using like "rate(counter_name[5m])" in web query page ***/
type promMeter struct {
	Label
	Metric *prometheus.CounterVec
}

func (p *promMeter) Update() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

func registerPromGauge(counterName string, labelName []string, labelValue []string) *promGauge {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: counterName,
	}, labelName)
	prometheus.MustRegister(gauge)

	return &promGauge{
		Label: Label{
			LabelName:  labelName,
			LabelValue: labelValue,
		},
		Metric: gauge,
	}
}

func registerPromMeter(counterName string, labelName []string, labelValue []string) *promMeter {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: counterName,
	}, labelName)
	prometheus.MustRegister(counter)

	return &promMeter{
		Label: Label{
			LabelName:  labelName,
			LabelValue: labelValue,
		},
		Metric: counter,
	}
}

func startPromHTTPServer() {
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	logrus.Fatal(http.ListenAndServe(":1988", nil))
}
