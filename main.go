package main

import (
	"log"

	"github.com/pegasus-kv/meta-proxy/rpc"
)

func main() {
	err := rpc.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
