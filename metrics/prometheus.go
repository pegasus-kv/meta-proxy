/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type label struct {
	LabelName  []string
	LabelValue []string
}

// promGauge metric for reporting the current total number which can be "add" or "delete"
type promGauge struct {
	label
	Metric *prometheus.GaugeVec
}

// Add add value of counter
func (p *promGauge) Add(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(float64(value))
}

// Inc add value of counter, value = 1
func (p *promGauge) Inc() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

// Decrease decrease value of counter
func (p *promGauge) Sub(value int64) {
	p.Metric.WithLabelValues(p.LabelValue...).Add(-float64(value))
}

// Dec decrease value of counter, value = 1
func (p *promGauge) Dec() {
	p.Metric.WithLabelValues(p.LabelValue...).Dec()
}

// promMeter metric for reporting the rate number which only can be "add"
// and the rate get by using like "rate(counter_name[5m])" in web query page ***/
type promMeter struct {
	label
	Metric *prometheus.CounterVec
}

// Update the counter, add 1
func (p *promMeter) Update() {
	p.Metric.WithLabelValues(p.LabelValue...).Inc()
}

func registerPromGauge(name string, labelName []string, labelValue []string) *promGauge {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
	}, labelName)
	prometheus.MustRegister(gauge)

	return &promGauge{
		label: label{
			LabelName:  labelName,
			LabelValue: labelValue,
		},
		Metric: gauge,
	}
}

func registerPromMeter(name string, labelName []string, labelValue []string) *promMeter {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
	}, labelName)
	prometheus.MustRegister(counter)

	return &promMeter{
		label: label{
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
	logrus.Fatal(http.ListenAndServe(":9091", nil))
}
