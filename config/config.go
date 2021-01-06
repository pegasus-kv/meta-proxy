package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ZookeeperOption struct {
	Address      []string `mapstructure:"address"`
	Root         string   `mapstructure:"root"`
	Timeout      int      `mapstructure:"timeout"`
	WatcherCount int      `mapstructure:"table_watcher_cache_capacity"`
}

type PerfCounterOption struct {
	Type string            `mapstructure:"type"`
	Tags map[string]string `mapstructure:"tags"`
}

var Config Configuration

type Configuration struct {
	ZookeeperOpt   ZookeeperOption   `mapstructure:"zookeeper"`
	PerfCounterOpt PerfCounterOption `mapstructure:"perfCounter"`
}

func InitConfig(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Panicf("unable find config file(meta-proxy.yml)")
		} else {
			logrus.Panicf("fatal error config file(meta-proxy.yml): %s", err)
		}
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		logrus.Panicf("unable to decode into struct, %s", err)
	}
}
