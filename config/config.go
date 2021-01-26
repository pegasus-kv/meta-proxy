package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// zookeeperOpts is the configuration for zookeeper client.
type zookeeperOpts struct {
	Address      []string `mapstructure:"address"`
	Root         string   `mapstructure:"root"`
	Timeout      int      `mapstructure:"timeout"`
	WatcherCount int      `mapstructure:"table_watcher_cache_capacity"`
}

// metricsOpts used for init the perfCounter type(now support the Falcon and Prometheus) and
type metricsOpts struct {
	Type string   `mapstructure:"type"`
	Tags []string `mapstructure:"tags"`
}

var GlobalConfig Configuration

// Configuration is the wrapper of zookeeperOpts and metricsOpts
type Configuration struct {
	ZookeeperOpts zookeeperOpts `mapstructure:"zookeeper"`
	MetricsOpts   metricsOpts   `mapstructure:"metric"`
}

// Init meta-proxy config using the config file
func Init(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Panicf("unable find config file \"%s\"", path)
		} else {
			logrus.Panicf("fatal error config file \"%s\":%s", path, err)
		}
	}

	err := viper.Unmarshal(&GlobalConfig)
	if err != nil {
		logrus.Panicf("unable to decode \"%s\" into struct: %s", path, err)
	}
	logrus.Infof("init config: %v", GlobalConfig)
}
