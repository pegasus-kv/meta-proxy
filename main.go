package main

import(
	 "github.com/pegasus-kv/meta-proxy/rpc"
)


func main() {
	server := rpc.NewServer()
	server.Serve()
}

