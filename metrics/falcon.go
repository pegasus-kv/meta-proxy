package metrics

import falcon "github.com/niean/goperfcounter"

type falconGauge struct {
	name string
}

// Add add value of counter
func (f *falconGauge) Add(value int64) {
	falcon.SetCounterCount(f.name, value)
}

// Inc add value of counter, value = 1
func (f *falconGauge) Inc() {
	f.Add(1)
}

// Decrease decrease value of counter
func (f *falconGauge) Decrease(value int64) {
	falcon.SetCounterCount(f.name, -value)
}

// Dec decrease value of counter, value = 1
func (f *falconGauge) Dec() {
	f.Decrease(1)
}

type falconMeter struct {
	name string
}

// Update add value of counter, value = 1
func (f *falconMeter) Update() {
	falcon.SetMeterCount(f.name, 1)
}

func registerFalconGauge(name string) *falconGauge {
	return &falconGauge{
		name: name,
	}
}

func registerFalconMeter(name string) *falconMeter {
	return &falconMeter{
		name: name,
	}
}
