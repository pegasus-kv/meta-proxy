package rpc

import (
	"log"
	"net"
	"net/rpc"
)

// Server is the RPC server that can handle rDSN-protocol RPC requests.
type Server struct {
}

// Serve blocks until the connection shutdown.
func (s *Server) Serve() {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:34601")
	if err != nil {
		log.Fatal(err)
		return
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print("connection accept:", err.Error())
			return
		}
		// TODO(wutao): add metric for connections number
		// TODO(wutao): add connections management

		// use one goroutine per connection
		go rpc.ServeCodec(newCodec(conn))
	}
}

// NewServer creates a meta-proxy server.
func NewServer() *Server {
	return &Server{}
}

// pegasusCodec implements the rDSN RPC protocol.
// It inherits from rpc.ServerCodec, each handles one TCP stream,
// so there's no concurrent call on it.
type pegasusCodec struct {
	conn net.Conn

	dec *decoder
}

func (c *pegasusCodec) ReadRequestHeader(req *rpc.Request) error {
	pegaReq, err := c.dec.readRequestMeta()
	if err != nil {
		return err
	}
	req.Seq = pegaReq.seqID
	// ServiceMethod requires to be in "Service.Method" format.
	req.ServiceMethod = "Thrift." + pegaReq.methodName
	return nil
}

func (c *pegasusCodec) ReadRequestBody(value interface{}) error {
	if value == nil {

	} else {

	}
	return nil
}

func (c *pegasusCodec) WriteResponse(*rpc.Response, interface{}) error {

}

func (c *pegasusCodec) Close() error {
	return c.conn.Close()
}

func newCodec(conn net.Conn) rpc.ServerCodec {
	return &pegasusCodec{
		conn: conn,

		dec: &decoder{
			reader: conn,
		},
	}
}
