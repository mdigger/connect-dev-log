package devlog

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
)

// Logger implements connect.Interceptor for RPC call logging.
type Logger struct {
	writer         io.Writer
	mu             sync.Mutex
	timeFormat     string
	showHeaders    bool
	protoFormatter ProtoFormatter
	pool           sync.Pool
}

// New creates a configured Logger instance.
// Writer must be thread-safe if shared across goroutines.
func New(w io.Writer, opts ...Option) *Logger {
	l := &Logger{
		writer:      w,
		timeFormat:  time.RFC3339Nano,
		showHeaders: false,
		protoFormatter: protojson.MarshalOptions{
			Multiline:     true,
			AllowPartial:  true,
			UseProtoNames: true,
		},
	}
	l.pool.New = func() any { return &strings.Builder{} }

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WrapUnary implements connect.UnaryInterceptorFunc.
func (l *Logger) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		buf := l.getBuilder()
		defer l.putBuilder(buf)

		l.logRequest(buf, req, start)
		resp, err := next(ctx, req)
		l.logResponse(buf, resp, err, time.Since(start))
		l.writeLog(buf)

		return resp, err
	}
}

// getBuilder gets a strings.Builder from pool.
func (l *Logger) getBuilder() *strings.Builder {
	return l.pool.Get().(*strings.Builder) //nolint:forcetypeassert,errcheck
}

// putBuilder returns builder to pool.
func (l *Logger) putBuilder(b *strings.Builder) {
	b.Reset()
	l.pool.Put(b)
}

// writeLog safely outputs the log.
func (l *Logger) writeLog(buf *strings.Builder) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.writer, buf.String())
}
