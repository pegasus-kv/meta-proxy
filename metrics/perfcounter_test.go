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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.Init("../meta-proxy.yml")
}

func TestParseTags(t *testing.T) {
	names, values := parseTags()
	assert.Contains(t, names, "region")
	assert.Contains(t, names, "service")
	assert.Contains(t, values, "c3tst_staging")
	assert.Contains(t, values, "meta_proxy")
}

func TestPrometheus(t *testing.T) {
	config.GlobalConfig.MetricsOpts.Type = "prometheus"
	gaugeCounter := RegisterGauge("promGaugeTest")
	meterCounter := RegisterMeter("promMeterTest")
	Init()
	// mock the promGauge counter: gaugeCounter = 0
	gaugeCounter.Add(100)
	gaugeCounter.Inc()
	gaugeCounter.Sub(100)
	gaugeCounter.Dec()

	// mock the promMeter: meterCounter = 1
	meterCounter.Update()
	time.Sleep(1000000000)
	resp, err := http.Get("http://localhost:9091/metrics")
	assert.Nil(t, err)
	defer resp.Body.Close()
	// the resp page content like: "counter value \n counter value \n"
	body, _ := ioutil.ReadAll(resp.Body)
	result := strings.Split(string(body), "\n")
	assert.Contains(t, result, "promGaugeTest{region=\"c3tst_staging\",service=\"meta_proxy\"} 0")
	assert.Contains(t, result, "promMeterTest{region=\"c3tst_staging\",service=\"meta_proxy\"} 1")
}

func TestFalcon(t *testing.T) {
	config.GlobalConfig.MetricsOpts.Type = "falcon"
	gaugeCounter := RegisterGauge("falconGaugeTest")
	meterCounter := RegisterMeter("falconMeterTest")
	Init()
	// mock the falconGauge counter
	gaugeCounter.Add(100)
	gaugeCounter.Inc()
	gaugeCounter.Sub(100)
	gaugeCounter.Dec()

	// mock the falconMeter
	meterCounter.Update()
}
