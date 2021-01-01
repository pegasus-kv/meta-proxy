package collector

import falcon "git.n.xiaomi.com/falcon-sdk/goperfcounter"

type FalconCounter struct {
	name string
}

func (p *FalconCounter) Incr() {
	falcon.SetCounterCount(p.name, 1)
}

func registerFalconCounter(counterName string) *FalconCounter {
	return &FalconCounter{
		name: counterName,
	}
}
