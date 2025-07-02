package devlog

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Logger implements connect.Interceptor for RPC call logging.
type Logger struct {
	writer         io.Writer
	mu             sync.Mutex
	timeFormat     string
	showHeaders    bool
	excludeHeaders map[string]bool
	protoFormatter func(m proto.Message) string
}

// Verify Logger fully implements connect.Interceptor.
var _ connect.Interceptor = (*Logger)(nil)

// New creates a configured Logger instance.
// The provided writer must be thread-safe if shared across goroutines.
func New(w io.Writer, opts ...Option) *Logger {
	l := &Logger{
		writer:      w,
		timeFormat:  time.RFC3339Nano,
		showHeaders: false,
		protoFormatter: protojson.MarshalOptions{
			Multiline:     true,
			UseProtoNames: true,
		}.Format,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WrapUnary implements connect.UnaryInterceptorFunc.
func (l *Logger) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()
		buf := getBuffer()
		defer putBuffer(buf)

		l.writeTimestamp(buf, start)

		buf.WriteString(req.Spec().Procedure)
		buf.WriteByte(' ')
		buf.WriteString(req.Spec().StreamType.String())
		buf.WriteByte(' ')
		buf.WriteString(req.Peer().Protocol)
		buf.WriteByte(' ')
		buf.WriteString(req.Peer().Addr)
		buf.WriteString("\n  * Unary request")

		if l.showHeaders {
			l.writeHeaders(buf, req.Header(), l.excludeHeaders)
		}

		if msg, ok := req.Any().(proto.Message); ok {
			buf.WriteString("\n  Request:")
			l.writeProto(buf, msg)
		}

		resp, err := next(ctx, req)
		stop := time.Since(start)

		if err != nil {
			l.writeError(buf, err)
		} else if msg, ok := resp.Any().(proto.Message); ok {
			buf.WriteString("\n  Response:")
			l.writeProto(buf, msg)
		}

		buf.WriteString("\n  completed in ")
		buf.WriteString(stop.String())
		buf.WriteByte('\n')
		l.writeOutput(buf)

		return resp, err
	}
}

func (l *Logger) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()
		buf := getBuffer()
		defer putBuffer(buf)

		l.writeTimestamp(buf, start)

		buf.WriteString(conn.Spec().Procedure)
		buf.WriteByte(' ')
		buf.WriteString(conn.Spec().StreamType.String())
		buf.WriteByte(' ')
		buf.WriteString(conn.Peer().Protocol)
		buf.WriteByte(' ')
		buf.WriteString(conn.Peer().Addr)
		buf.WriteString("\n  * Start streaming")

		if l.showHeaders {
			l.writeHeaders(buf, conn.RequestHeader(), l.excludeHeaders)
		}

		buf.WriteByte('\n')
		l.writeOutput(buf)
		buf.Reset()

		wrappedConn := &streamingHandlerConn{
			StreamingHandlerConn: conn,
			logger:               l,
		}

		err := next(ctx, wrappedConn)
		stop := time.Since(start)

		l.writeTimestamp(buf, time.Now())
		buf.WriteString(conn.Spec().Procedure)
		buf.WriteByte(' ')
		buf.WriteString(conn.Spec().StreamType.String())
		buf.WriteByte(' ')
		buf.WriteString(conn.Peer().Protocol)
		buf.WriteByte(' ')
		buf.WriteString(conn.Peer().Addr)
		buf.WriteString("\n  * Stop streaming")

		if err != nil {
			l.writeError(buf, err)
		}

		buf.WriteString("\n  completed in ")
		buf.WriteString(stop.String())
		i := wrappedConn.received.Load()
		buf.WriteString(" (received: ")
		buf.WriteString(strconv.FormatInt(i, 10))
		i = wrappedConn.sended.Load()
		buf.WriteString(", sended: ")
		buf.WriteString(strconv.FormatInt(i, 10))
		buf.WriteString(")\n")
		l.writeOutput(buf)

		return err
	}
}

func (l *Logger) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	// TODO: not implemented
	return next
}

func (l *Logger) writeOutput(buf *bytes.Buffer) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := buf.WriteTo(l.writer)
	return err
}
