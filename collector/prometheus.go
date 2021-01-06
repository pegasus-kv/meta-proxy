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

/*** PromGauge metric for reporting the current total number which can be "add" or "delete" ***/
type PromGauge struct {
	Label
	Metric *prometheus.GaugeVec
}

func (p *PromGauge) Add(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(float64(value))
}

func (p *PromGauge) Incr() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

func (p *PromGauge) Delete(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(-float64(value))
}

func (p *PromGauge) Decrease() {
	p.Metric.WithLabelValues(p.LabelValue...).Dec()
}

/*** PromMeter metric for reporting the rate number which only can be "add" ***/
/***  and the rate get by using like "rate(counter_name[5m])" in web query page ***/
type PromMeter struct {
	Label
	Metric *prometheus.CounterVec
}

func (p *PromMeter) Update() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

func registerPromGauge(counterName string, labelName []string, labelValue []string) *PromGauge {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: counterName,
	}, labelName)
	prometheus.MustRegister(gauge)

	return &PromGauge{
		Label: Label{
			LabelName:  labelName,
			LabelValue: labelValue,
		},
		Metric: gauge,
	}
}

func registerPromMeter(counterName string, labelName []string, labelValue []string) *PromMeter {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: counterName,
	}, labelName)
	prometheus.MustRegister(counter)

	return &PromMeter{
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
	logrus.Fatal(http.ListenAndServe(":1988", nil))
}
