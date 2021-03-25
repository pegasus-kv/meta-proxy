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
	"fmt"

	falcon "github.com/niean/goperfcounter"
	"github.com/sirupsen/logrus"
)

type falconGauge struct {
	counterName string
	tagsName    []string
}

// Add add value of counter
func (f *falconGauge) Add(value int64) {
	f.AddWithTags([]string{}, value)
}

// Add add value of counter with custom tags
func (f *falconGauge) AddWithTags(tagsValue []string, counterValue int64) {
	falcon.SetCounterCount(parseToCounterName(f.counterName, f.tagsName, tagsValue), counterValue)
}

// Inc add value of counter, value = 1
func (f *falconGauge) Inc() {
	f.IncWithTags([]string{})
}

// Inc add value of counter with custom tags, value = 1
func (f *falconGauge) IncWithTags(tagsValue []string) {
	f.AddWithTags(tagsValue, 1)
}

// Decrease decrease value of counter
func (f *falconGauge) Sub(value int64) {
	f.SubWithTags([]string{}, value)
}

// Decrease decrease value of counter with custom tags
func (f *falconGauge) SubWithTags(tagsValue []string, counterValue int64) {
	falcon.SetCounterCount(parseToCounterName(f.counterName, f.tagsName, tagsValue), -counterValue)
}

// Dec decrease value of counter, value = 1
func (f *falconGauge) Dec() {
	f.DecWithTags([]string{})
}

// Dec decrease value of counter with custom tags, value = 1
func (f *falconGauge) DecWithTags(tagsValue []string) {
	f.SubWithTags(tagsValue, 1)
}

type falconMeter struct {
	counterName string
	tagsName    []string
}

// Update add value of counter, value = 1
func (f *falconMeter) Update() {
	f.UpdateWithTags([]string{})
}

// UpdateWithTags add value of counter with custom tags, value = 1
func (f *falconMeter) UpdateWithTags(tags []string) {
	falcon.SetMeterCount(parseToCounterName(f.counterName, f.tagsName, tags), 1)
}

func registerFalconGauge(counterName string, tagsName []string) *falconGauge {
	return &falconGauge{
		counterName: counterName,
		tagsName:    tagsName,
	}
}

func registerFalconMeter(counterName string, tagsName []string) *falconMeter {
	return &falconMeter{
		counterName: counterName,
		tagsName:    tagsName,
	}
}

// transfer counterName registered and tags into final counterName
func parseToCounterName(counterName string, tagsName []string, tagsValue []string) string {
	tagsName = combineConfigTagsName(tagsName)
	tagsValue = combineConfigTagsValue(tagsValue)
	if len(tagsName) != len(tagsValue) {
		logrus.Panicf("[%s] tag's length is invalid: tagsName=%s, tagsValue=%s", counterName, tagsName, tagsValue)
	}

	for n := range tagsName {
		counterName = fmt.Sprintf("%s,%s=%s", counterName, tagsName[n], tagsValue[n])
	}
	return counterName
}
