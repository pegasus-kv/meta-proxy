package rpc

import (
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
	dec := &decoder{
		reader: conn,
	}

	for {
		req, err := dec.readRequest()
		if err != nil {
			if err != io.EOF {
				continue
			}
			break
		}

		go func() {
			result := req.handler(req.args)
			sendResponse(result)
		}()
	}
}

func sendResponse(result ResponseResult) error {
	return nil
}
