package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// promGauge tags for reporting the current total number which can be "add" or "delete"
type promGauge struct {
	labelsName []string
	metric     *prometheus.GaugeVec
}

// Add add value of counter
func (p *promGauge) Add(value int64) {
	p.AddWithTags([]string{}, value)
}

// Add add value of counter with custom tags
func (p *promGauge) AddWithTags(tagsValue []string, counterValue int64) {
	p.metric.WithLabelValues(combineConfigTagsValue(tagsValue)...).Add(float64(counterValue))
}

// Inc add value of counter, value = 1
func (p *promGauge) Inc() {
	p.IncWithTags([]string{})
}

// Inc add value of counter with custom tags, value = 1
func (p *promGauge) IncWithTags(tagsValue []string) {
	p.metric.WithLabelValues(combineConfigTagsValue(tagsValue)...).Inc()
}

// Decrease decrease value of counter
func (p *promGauge) Sub(value int64) {
	p.SubWithTags([]string{}, value)
}

// Decrease decrease value of counter with custom tags
func (p *promGauge) SubWithTags(tagsValue []string, counterValue int64) {
	p.metric.WithLabelValues(combineConfigTagsValue(tagsValue)...).Add(-float64(counterValue))
}

// Dec decrease value of counter, value = 1
func (p *promGauge) Dec() {
	p.DecWithTags([]string{})
}

// Dec decrease value of counter with custom tags, value = 1
func (p *promGauge) DecWithTags(tagsValue []string) {
	p.metric.WithLabelValues(combineConfigTagsValue(tagsValue)...).Dec()
}

// promMeter promMeter for reporting the rate number which only can be "add"
// and the rate get by using like "rate(counter_name[5m])" in web query page ***/
type promMeter struct {
	labelsName []string
	metric     *prometheus.CounterVec
}

// Update the counter, add 1
func (p *promMeter) Update() {
	p.UpdateWithTags([]string{})
}

// Update the counter with custom tags, add 1
func (p *promMeter) UpdateWithTags(tags []string) {
	p.metric.WithLabelValues(combineConfigTagsValue(tags)...).Inc()
}

func registerPromGauge(counterName string, labelsName []string) *promGauge {
	labelsName = combineConfigTagsName(labelsName)
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: counterName,
	}, labelsName)
	prometheus.MustRegister(gauge)

	return &promGauge{
		labelsName: labelsName,
		metric:     gauge,
	}
}

func registerPromMeter(counterName string, labelsName []string) *promMeter {
	labelsName = combineConfigTagsName(labelsName)
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: counterName,
	}, labelsName)
	prometheus.MustRegister(counter)

	return &promMeter{
		labelsName: labelsName,
		metric:     counter,
	}
}

func startPromHTTPServer() {
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	logrus.Fatal(http.ListenAndServe(":9091", nil))
}
