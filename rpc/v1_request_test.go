package rpc

import (
	"encoding/binary"

	"github.com/XiaoMi/pegasus-go-client/session"
	"github.com/apache/thrift/lib/go/thrift"
)

// pegasusV1Codec implements pegasus-go-client.rpc.Codec
type pegasusV1Codec struct {
}

const pegasusV1RequestHeader = 16

func (p *pegasusV1Codec) Marshal(v interface{}) ([]byte, error) {
	r, _ := v.(*session.PegasusRpcCall)

	thriftMeta := &ThriftRequestMetaV1{
		AppID:          &r.Gpid.Appid,
		PartitionIndex: &r.Gpid.PartitionIndex,
	}
	buf := thrift.NewTMemoryBuffer()
	buf.Write(make([]byte, pegasusV1RequestHeader)) // placeholder for header
	oprot := thrift.NewTBinaryProtocolTransport(buf)
	_ = thriftMeta.Write(oprot)
	metaLength := buf.Len() - pegasusV1RequestHeader

	// write body
	var err error
	if err = oprot.WriteMessageBegin(r.Name, thrift.CALL, r.SeqId); err != nil {
		return nil, err
	}
	if err = r.Args.Write(oprot); err != nil {
		return nil, err
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return nil, err
	}
	bodyLength := buf.Len() - metaLength - pegasusV1RequestHeader

	// write header
	headerBuf := buf.Bytes()
	copy(headerBuf[0:4], []byte{'T', 'H', 'F', 'T'})
	binary.BigEndian.PutUint32(headerBuf[4:8], uint32(1))
	binary.BigEndian.PutUint32(headerBuf[8:12], uint32(metaLength))
	binary.BigEndian.PutUint32(headerBuf[12:16], uint32(bodyLength))

	return buf.Bytes(), nil
}

func (p *pegasusV1Codec) Unmarshal(data []byte, v interface{}) error {
	// unimplemented
	return nil
}

func (p *pegasusV1Codec) String() string {
	return "pegasusv1"
}
