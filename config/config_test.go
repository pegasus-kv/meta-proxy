package config

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestConfig(t *testing.T) {
	Init("yaml/meta-proxy-example.yml")
	config := Configuration{
		ZookeeperOpts: zookeeperOpts{
			Address:      []string{"127.0.0.1:22181", "127.0.0.2:22181"},
			Root:         "/pegasus-cluster",
			Timeout:      1000,
			WatcherCount: 1024,
		},
		MetricsOpts: metricsOpts{
			Type: "falcon",
			Tags: []string{"region=local_tst", "service=meta_proxy"},
		},
	}

	assert.Equal(t, config, GlobalConfig)
}
