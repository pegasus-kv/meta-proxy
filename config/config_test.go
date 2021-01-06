package config

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestConfig(t *testing.T) {
	InitConfig("../meta-proxy.yml")
	config := Configuration{
		Zk: Zookeeper{
			Address:      []string{"127.0.0.1:22181", "127.0.0.2:22181"},
			Root:         "/pegasus-cluster",
			Timeout:      1000,
			WatcherCount: 1024,
		},
		Pfc: PerfCounter{
			Type: "falcon",
			Tags: map[string]string{
				"region":  "c3tst_staging",
				"service": "meta_proxy",
			},
		},
	}

	assert.Equal(t, config, Cfg)
}
