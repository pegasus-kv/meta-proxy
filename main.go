package main

import (
	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/pegasus-kv/meta-proxy/meta"
	"github.com/pegasus-kv/meta-proxy/metrics"
	"github.com/pegasus-kv/meta-proxy/rpc"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func main() {
	config.Init(os.Args[1])
	logrus.SetOutput(&lumberjack.Logger{
		Filename:  "meta-proxy.log",
		MaxSize:   500, // MB
		MaxAge:    7,   // days
		LocalTime: true,
	})

	meta.Init()
	err := rpc.Serve()
	if err != nil {
		logrus.Fatalf("start server error: %s", err)
	}

	metrics.Init() // metrics must init at last to make sure other package metric counter register completed
}
