package devlog

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

// WrapStreamingClient implements connect.StreamingClientInterceptor.
// It logs the start of client streaming calls and returns a wrapped connection.
func (l *Logger) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		start := time.Now()
		buf := getBuilder()

		l.logStreamStart(buf, "Client stream started", spec, start)
		l.writeLog(buf)
		putBuilder(buf)

		return &streamingClientConn{
			StreamingClientConn: next(ctx, spec),
			logger:              l,
			start:               start,
		}
	}
}

// WrapStreamingHandler implements connect.StreamingHandlerInterceptor.
// It logs the start and end of handler streaming calls including headers if enabled.
func (l *Logger) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()
		buf := getBuilder()
		defer putBuilder(buf)

		l.logStreamStart(buf, "Handler stream started", conn.Spec(), start)

		if l.showHeaders {
			l.logHeaders(buf, "Request headers", conn.RequestHeader())
		}

		err := next(ctx, &streamingHandlerConn{
			StreamingHandlerConn: conn,
			logger:               l,
		})
		l.logStreamEnd(buf, err, time.Since(start))
		l.writeLog(buf)

		return err
	}
}

// streamingClientConn wraps connect.StreamingClientConn to add logging capabilities.
type streamingClientConn struct {
	connect.StreamingClientConn
	logger *Logger
	start  time.Time
	mu     sync.Mutex
}

// Send logs the message or error before delegating to the underlying connection.
func (s *streamingClientConn) Send(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingClientConn.Send(msg)
	if err != nil || (msg != nil && s.logger.protoFormatter != nil) {
		buf := getBuilder()
		defer putBuilder(buf)

		if err != nil {
			buf.WriteString("[")
			buf.WriteString(time.Now().Format(s.logger.timeFormat))
			buf.WriteString("] Send error: ")
			buf.WriteString(err.Error())
			buf.WriteByte('\n')
		} else if m, ok := msg.(proto.Message); ok {
			s.logger.logProtoMessage(buf, "Sent message", m)
		}

		s.logger.writeLog(buf)
	}

	return err
}

// Receive logs the received message or error before delegating to the underlying connection.
func (s *streamingClientConn) Receive(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingClientConn.Receive(msg)
	if err != nil || (msg != nil && s.logger.protoFormatter != nil) {
		buf := getBuilder()
		defer putBuilder(buf)

		if err != nil {
			buf.WriteString("[")
			buf.WriteString(time.Now().Format(s.logger.timeFormat))

			if errors.Is(err, io.EOF) {
				buf.WriteString("] Stream closed by server\n")
			} else {
				buf.WriteString("] Receive error: ")
				buf.WriteString(err.Error())
				buf.WriteByte('\n')
			}
		} else if m, ok := msg.(proto.Message); ok {
			s.logger.logProtoMessage(buf, "Received message", m)
		}

		s.logger.writeLog(buf)
	}

	return err
}

// CloseRequest logs any error that occurs during request closing.
func (s *streamingClientConn) CloseRequest() error {
	err := s.StreamingClientConn.CloseRequest()
	if err != nil {
		buf := getBuilder()
		defer putBuilder(buf)

		buf.WriteString("[")
		buf.WriteString(time.Now().Format(s.logger.timeFormat))
		buf.WriteString("] CloseRequest error: ")
		buf.WriteString(err.Error())
		buf.WriteByte('\n')
		s.logger.writeLog(buf)
	}
	return err
}

// CloseResponse logs any error that occurs during response closing.
func (s *streamingClientConn) CloseResponse() error {
	err := s.StreamingClientConn.CloseResponse()
	if err != nil {
		buf := getBuilder()
		defer putBuilder(buf)

		buf.WriteString("[")
		buf.WriteString(time.Now().Format(s.logger.timeFormat))
		buf.WriteString("] CloseResponse error: ")
		buf.WriteString(err.Error())
		buf.WriteByte('\n')
		s.logger.writeLog(buf)
	}
	return err
}

// streamingHandlerConn wraps connect.StreamingHandlerConn to add logging capabilities.
type streamingHandlerConn struct {
	connect.StreamingHandlerConn
	logger *Logger
	mu     sync.Mutex
}

// Send logs the message or error before delegating to the underlying connection.
func (s *streamingHandlerConn) Send(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingHandlerConn.Send(msg)
	if err != nil || (msg != nil && s.logger.protoFormatter != nil) {
		buf := getBuilder()
		defer putBuilder(buf)

		if err != nil {
			buf.WriteString("[")
			buf.WriteString(time.Now().Format(s.logger.timeFormat))
			buf.WriteString("] Send error: ")
			buf.WriteString(err.Error())
			buf.WriteByte('\n')
		} else if m, ok := msg.(proto.Message); ok {
			s.logger.logProtoMessage(buf, "Sent message", m)
		}

		s.logger.writeLog(buf)
	}

	return err
}

// Receive logs the received message or error before delegating to the underlying connection.
func (s *streamingHandlerConn) Receive(msg any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.StreamingHandlerConn.Receive(msg)
	if err != nil || (msg != nil && s.logger.protoFormatter != nil) {
		buf := getBuilder()
		defer putBuilder(buf)

		if err != nil {
			buf.WriteString("[")
			buf.WriteString(time.Now().Format(s.logger.timeFormat))

			if errors.Is(err, io.EOF) {
				buf.WriteString("] Stream closed by client\n")
			} else {
				buf.WriteString("] Receive error: ")
				buf.WriteString(err.Error())
				buf.WriteByte('\n')
			}
		} else if m, ok := msg.(proto.Message); ok {
			s.logger.logProtoMessage(buf, "Received message", m)
		}

		s.logger.writeLog(buf)
	}

	return err
}
