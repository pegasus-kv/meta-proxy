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

func parseToCounterName(counterName string, tagsName []string, tagsValue []string) string {
	if len(tagsName) == 0 {
		return counterName
	}

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
