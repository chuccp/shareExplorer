package quic

import (
	"github.com/chuccp/utils/io"
	"github.com/chuccp/utils/log"
	"github.com/lucas-clemente/quic-go"
)

type conn struct {
	stream quic.Stream
	readStream *io.ReadStream
	writeStream *io.WriteStream
}

func newConn(stream quic.Stream) *conn {
	stream.Write([]byte{})
	return &conn{stream: stream,readStream:io.NewReadStream(stream),writeStream:io.NewWriteStream(stream)}
}
func (conn *conn) Start() {
	for{
		readByte, err := conn.readStream.ReadLine()
		if err != nil {
			return
		}
		log.Info(string(readByte))
	}
}