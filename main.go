package main

import (
	"log"

	"github.com/pegasus-kv/meta-proxy/meta"
	"github.com/pegasus-kv/meta-proxy/rpc"
)

func main() {

	meta.Init()

	err := rpc.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
