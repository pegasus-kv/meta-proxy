package main

import (
	"github.com/pegasus-kv/meta-proxy/collector"
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/pegasus-kv/meta-proxy/meta"
	"github.com/pegasus-kv/meta-proxy/rpc"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func main() {
	logrus.SetOutput(&lumberjack.Logger{
		Filename:  "meta-proxy.log",
		MaxSize:   500, // MB
		MaxAge:    7,   // days
		LocalTime: true,
	})
	config.InitConfig(os.Args[0])
	collector.InitPerfCounter()

	meta.Init()
	err := rpc.Serve()
	if err != nil {
		logrus.Fatalf("start server error: %s", err)
	}
}
