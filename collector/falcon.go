package collector

import falcon "git.n.xiaomi.com/falcon-sdk/goperfcounter"

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
