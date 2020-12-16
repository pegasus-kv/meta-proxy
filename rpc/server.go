package rpc

import (
	"io"
	"log"
	"net"
	"net/rpc"
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
		go rpc.ServeCodec(newCodec(conn))
	}
}

// pegasusCodec implements the rDSN RPC protocol.
// It inherits from rpc.ServerCodec, each handles one TCP stream,
// so there's no concurrent call on it.
type pegasusCodec struct {
	conn io.ReadWriteCloser

	dec *decoder
}

func (c *pegasusCodec) ReadRequestHeader(req *rpc.Request) error {
	pegaReq, err := c.dec.readRequest()
	if err != nil {
		return err
	}
	req.Seq = pegaReq.seqID
	// ServiceMethod requires to be in "Service.Method" format.
	req.ServiceMethod = "Thrift." + pegaReq.methodName
	return nil
}

func (c *pegasusCodec) ReadRequestBody(value interface{}) error {
	return nil
}

func (c *pegasusCodec) WriteResponse(*rpc.Response, interface{}) error {
	return nil
}

func (c *pegasusCodec) Close() error {
	return c.conn.Close()
}

// conn is a network connection but abstracted as a ReadWriteCloser here in order to do mock test.
func newCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &pegasusCodec{
		conn: conn,

		dec: &decoder{
			reader: conn,
		},
	}
}
