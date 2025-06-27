package devlog

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

// WrapStreamingClient implements connect.StreamingClientInterceptor.
func (l *Logger) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		start := time.Now()
		buf := l.getBuilder()
		l.logStreamStart(buf, "Client stream started", spec, start)

		return &streamingClientConn{
			StreamingClientConn: next(ctx, spec),
			logger:              l,
			buf:                 buf,
			start:               start,
		}
	}
}

// WrapStreamingHandler implements connect.StreamingHandlerInterceptor.
func (l *Logger) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()

		buf := l.getBuilder()
		defer l.putBuilder(buf)

		l.logStreamStart(buf, "Handler stream started", conn.Spec(), start)

		if l.showHeaders {
			l.logHeaders(buf, "Request headers", conn.RequestHeader())
		}

		err := next(ctx, &streamingHandlerConn{
			StreamingHandlerConn: conn,
			logger:               l,
			buf:                  buf,
		})
		l.logStreamEnd(buf, err, time.Since(start))
		l.writeLog(buf)

		return err
	}
}

type streamingClientConn struct {
	connect.StreamingClientConn
	logger *Logger
	buf    *strings.Builder
	start  time.Time
	mu     sync.Mutex
}

func (s *streamingClientConn) Send(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingClientConn.Send(msg)
	if err != nil {
		s.buf.WriteString("[")
		s.buf.WriteString(time.Now().Format(s.logger.timeFormat))
		s.buf.WriteString("] Send error: ")
		s.buf.WriteString(err.Error())
		s.buf.WriteByte('\n')
	} else if m, ok := msg.(proto.Message); ok {
		s.logger.logProtoMessage(s.buf, "Sent message", m)
	}

	return err
}

func (s *streamingClientConn) Receive(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingClientConn.Receive(msg)
	if err != nil {
		s.buf.WriteString("[")
		s.buf.WriteString(time.Now().Format(s.logger.timeFormat))

		if errors.Is(err, io.EOF) {
			s.buf.WriteString("] Stream closed by server\n")
		} else {
			s.buf.WriteString("] Receive error: ")
			s.buf.WriteString(err.Error())
			s.buf.WriteByte('\n')
		}
	} else if m, ok := msg.(proto.Message); ok {
		s.logger.logProtoMessage(s.buf, "Received message", m)
	}

	return err
}

type streamingHandlerConn struct {
	connect.StreamingHandlerConn
	logger *Logger
	buf    *strings.Builder
	mu     sync.Mutex
}

func (s *streamingHandlerConn) Send(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingHandlerConn.Send(msg)
	if err != nil {
		s.buf.WriteString("[")
		s.buf.WriteString(time.Now().Format(s.logger.timeFormat))
		s.buf.WriteString("] Send error: ")
		s.buf.WriteString(err.Error())
		s.buf.WriteByte('\n')
	} else if m, ok := msg.(proto.Message); ok {
		s.logger.logProtoMessage(s.buf, "Sent message", m)
	}

	return err
}

func (s *streamingHandlerConn) Receive(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingHandlerConn.Receive(msg)
	if err != nil {
		s.buf.WriteString("[")
		s.buf.WriteString(time.Now().Format(s.logger.timeFormat))

		if errors.Is(err, io.EOF) {
			s.buf.WriteString("] Stream closed by client\n")
		} else {
			s.buf.WriteString("] Receive error: ")
			s.buf.WriteString(err.Error())
			s.buf.WriteByte('\n')
		}
	} else if m, ok := msg.(proto.Message); ok {
		s.logger.logProtoMessage(s.buf, "Received message", m)
	}

	return err
}
