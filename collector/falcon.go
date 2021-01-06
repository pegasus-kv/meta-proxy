package collector

import falcon "github.com/niean/goperfcounter"

type FalconCounter struct {
	name string
}

func (f *FalconCounter) Add(value int64) {
	falcon.SetCounterCount(f.name, value)
}

func (f *FalconCounter) Incr() {
	f.Add(1)
}

func registerFalconCounter(counterName string) *FalconCounter {
	return &FalconCounter{
		name: counterName,
	}
}
