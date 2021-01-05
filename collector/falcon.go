package collector

import falcon "git.n.xiaomi.com/falcon-sdk/goperfcounter"

type FalconCounter struct {
	name string
}

func (f *FalconCounter) Incr() {
	falcon.SetCounterCount(f.name, 1)
}

func registerFalconCounter(counterName string) *FalconCounter {
	return &FalconCounter{
		name: counterName,
	}
}
