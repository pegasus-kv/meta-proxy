package collector

import falcon "github.com/niean/goperfcounter"

type FalconMetric struct {
	name string
}

/*** Falcon Gauge metric API for reporting the current total number which can be "add" or "delete" ***/
func (f *FalconMetric) Add(value int64) {
	falcon.SetCounterCount(f.name, value)
}

func (f *FalconMetric) Incr() {
	f.Add(1)
}

func (f *FalconMetric) Delete(value int64) {
	falcon.SetCounterCount(f.name, -value)
}

func (f *FalconMetric) Decrease() {
	f.Delete(1)
}

/*** Falcon Meter metric API for reporting the rate number which only can be "add" ***/
func (f *FalconMetric) Update() {
	falcon.SetMeterCount(f.name, 1)
}

func registerFalconMetric(counterName string) *FalconMetric {
	return &FalconMetric{
		name: counterName,
	}
}
