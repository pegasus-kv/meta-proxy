package main

import (
	"os"

	"github.com/pegasus-kv/meta-proxy/config"
	"github.com/pegasus-kv/meta-proxy/meta"
	"github.com/pegasus-kv/meta-proxy/metrics"
	"github.com/pegasus-kv/meta-proxy/rpc"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	logrus.SetOutput(&lumberjack.Logger{
		Filename:  os.Args[2],
		MaxSize:   500, // MB
		MaxAge:    7,   // days
		LocalTime: true,
	})

	config.Init(os.Args[1])
	meta.Init()
	err := rpc.Serve()
	if err != nil {
		logrus.Fatalf("start server error: %s", err)
	}
	metrics.Init() // metrics must init at last to make sure other package metric counter register completed
}
