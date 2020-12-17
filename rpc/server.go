package rpc

import (
	"context"
	"io"
	"log"
	"net"
)

// Serve blocks until the connection shutdown.
func Serve() error {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:34601")
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print("connection accept:", err.Error())
			continue
		}
		// TODO(wutao): add metric for connections number
		// TODO(wutao): add connections management

		// use one goroutine per connection
		go serveConn(conn)
	}
}

// conn is a network connection but abstracted as a ReadWriteCloser here in order to do mock test.
// The caller typically invokes serveConn in a go statement.
func serveConn(conn io.ReadWriteCloser) {
	dec := &requestDecoder{
		reader: conn,
	}
	enc := &responseEncoder{
		writer: conn,
	}

	ctx, cancel := context.WithCancel(context.Background())
	for {
		req, err := dec.readRequest()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				// TODO(wutao): send back rpc response for this error if the request is fully read
				continue
			}
			break
		}

		go func() {
			result := req.handler(ctx, req.args)
			err := enc.sendResponse(req, result)
			if err != nil {
				log.Println(err)
			}
		}()
	}

	// cancel the ongoing requests
	cancel()
	<-ctx.Done()
}
