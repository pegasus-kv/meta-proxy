package rpc

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

type responseEncoder struct {
	writer io.Writer

	mu sync.Mutex
}

// sendResponse is a thread-safe wrapper of doSendResponse.
func (e *responseEncoder) sendResponse(req *pegasusRequest, result ResponseResult) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.doSendResponse(req, result)
}

func (e *responseEncoder) doSendResponse(req *pegasusRequest, result ResponseResult) error {
	// prepare response bytes

	buf := thrift.NewTMemoryBuffer()
	oprot := thrift.NewTBinaryProtocolTransport(buf)
	err := oprot.WriteI32(0) // response length is still unknown now.
	if err != nil {
		return err
	}

	// error code
	if err = oprot.WriteString("ERR_OK"); err != nil {
		return err
	}

	// write response
	if err = oprot.WriteMessageBegin(req.methodName+"_ACK", thrift.REPLY, int32(req.seqID)); err != nil {
		return err
	}
	if err = result.Write(oprot); err != nil {
		return err
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return err
	}

	// response length is now got
	respLen := buf.Len()
	binary.BigEndian.PutUint32(buf.Bytes(), uint32(respLen))

	if _, err := e.writer.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
