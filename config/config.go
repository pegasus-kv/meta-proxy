package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Zookeeper struct {
	Address      []string `mapstructure:"address"`
	Root         string   `mapstructure:"root"`
	Timeout      int      `mapstructure:"timeout"`
	WatcherCount int      `mapstructure:"watcher_count"`
}

type PerfCounter struct {
	Type string            `mapstructure:"type"`
	Tags map[string]string `mapstructure:"tags"`
}

var Cfg Configuration

type Configuration struct {
	Zk  Zookeeper   `mapstructure:"Zookeeper"`
	Pfc PerfCounter `mapstructure:"PerfCounter"`
}

func init() {
	viper.SetConfigFile("meta-proxy.yml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Panicf("unable find config file(meta-proxy.yml)")
		} else {
			logrus.Panicf("fatal error config file(meta-proxy.yml): %s", err)
		}
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		logrus.Panicf("unable to decode into struct, %s", err)
	}
}
