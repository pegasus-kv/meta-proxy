package config

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestConfig(t *testing.T) {
	InitConfig("../meta-proxy.yml")
	config := Configuration{
		Zk: Zookeeper{
			Address:      []string{"127.0.0.1:22181", "127.0.0.2:22181"},
			Root:         "pegasus-cluster",
			Timeout:      1000,
			WatcherCount: 1024,
		},
		Pfc: PerfCounter{
			Type: "prometheus",
			Tags: map[string]string{
				"region":  "c3tst_staging",
				"service": "meta_proxy",
			},
		},
	}

	assert.Equal(t, config, Cfg)
}
