package config

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestConfig(t *testing.T) {
	InitConfig("../meta-proxy.yml")
	config := Configuration{
		ZookeeperOpt: ZookeeperOption{
			Address:      []string{"127.0.0.1:22181", "127.0.0.2:22181"},
			Root:         "/pegasus-cluster",
			Timeout:      1000,
			WatcherCount: 1024,
		},
		PerfCounterOpt: PerfCounterOption{
			Type: "falcon",
			Tags: map[string]string{
				"region":  "c3tst_staging",
				"service": "meta_proxy",
			},
		},
	}

	assert.Equal(t, config, Config)
}
