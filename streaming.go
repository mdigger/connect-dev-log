package devlog

import (
	"errors"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

type streamingHandlerConn struct {
	connect.StreamingHandlerConn
	logger           *Logger
	received, sended atomic.Int64
}

func (s *streamingHandlerConn) Send(msg any) error {
	err := s.StreamingHandlerConn.Send(msg)
	if err != nil || msg != nil {
		buf := getBuffer()
		defer putBuffer(buf)

		if s.logger.timeFormat != "" {
			writeTimestamp(buf, time.Now().Format(s.logger.timeFormat))
		}

		buf.WriteString(s.Spec().Procedure)
		buf.WriteByte(' ')
		buf.WriteString(s.Spec().StreamType.String())
		buf.WriteByte(' ')
		buf.WriteString(s.Peer().Protocol)
		buf.WriteByte(' ')
		buf.WriteString(s.Peer().Addr)

		if err != nil {
			writeError(buf, err)
		} else if m, ok := msg.(proto.Message); ok {
			i := s.sended.Add(1)
			buf.WriteString("\n  Sent message ")
			buf.WriteString(strconv.FormatInt(i, 10))
			buf.WriteByte(':')
			writeProto(buf, s.logger.protoFormatter(m))
		}

		buf.WriteByte('\n')
		s.logger.writeOutput(buf)
	}

	return err
}

func (s *streamingHandlerConn) Receive(msg any) error {
	err := s.StreamingHandlerConn.Receive(msg)
	if err != nil || msg != nil {
		buf := getBuffer()
		defer putBuffer(buf)

		if s.logger.timeFormat != "" {
			writeTimestamp(buf, time.Now().Format(s.logger.timeFormat))
		}

		buf.WriteString(s.Spec().Procedure)
		buf.WriteByte(' ')
		buf.WriteString(s.Spec().StreamType.String())
		buf.WriteByte(' ')
		buf.WriteString(s.Peer().Protocol)
		buf.WriteByte(' ')
		buf.WriteString(s.Peer().Addr)

		if err != nil {
			if errors.Is(err, io.EOF) {
				buf.WriteString("\n  Receive stream closed by client")
			} else {
				writeError(buf, err)
			}
		} else if m, ok := msg.(proto.Message); ok {
			i := s.received.Add(1)
			buf.WriteString("\n  Received message ")
			buf.WriteString(strconv.FormatInt(i, 10))
			buf.WriteByte(':')
			writeProto(buf, s.logger.protoFormatter(m))
		}

		buf.WriteByte('\n')
		s.logger.writeOutput(buf)
	}

	return err
}
