package devlog

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

// logRequest logs incoming request details including procedure, peer info and headers.
func (l *Logger) logRequest(buf *strings.Builder, req connect.AnyRequest, start time.Time) {
	buf.WriteString("[")
	buf.WriteString(start.Format(l.timeFormat))
	buf.WriteString("] ")
	buf.WriteString(req.Spec().Procedure)
	buf.WriteString(" from ")
	buf.WriteString(req.Peer().Addr)
	buf.WriteByte('\n')

	buf.WriteString("  StreamType: ")
	buf.WriteString(req.Spec().StreamType.String())
	buf.WriteString("\n  HTTPMethod: ")
	buf.WriteString(req.HTTPMethod())
	buf.WriteByte('\n')

	if l.showHeaders {
		l.logHeaders(buf, "Request headers", req.Header())
	}

	if msg, ok := req.Any().(proto.Message); ok {
		l.logProtoMessage(buf, "Request message", msg)
	}
}

// logResponse logs the response or error with duration information.
func (l *Logger) logResponse(buf *strings.Builder, resp connect.AnyResponse, err error, duration time.Duration) {
	buf.WriteString("[")
	buf.WriteString(time.Now().Format(l.timeFormat))
	buf.WriteString("] Completed in ")
	buf.WriteString(duration.String())
	buf.WriteByte('\n')

	if err != nil {
		l.logError(buf, err)
	} else if msg, ok := resp.Any().(proto.Message); ok {
		l.logProtoMessage(buf, "Response message", msg)
	}
}

// logStreamStart records stream initialization details.
func (l *Logger) logStreamStart(buf *strings.Builder, title string, spec connect.Spec, start time.Time) {
	buf.WriteString("[")
	buf.WriteString(start.Format(l.timeFormat))
	buf.WriteString("] ")
	buf.WriteString(title)
	buf.WriteString(": ")
	buf.WriteString(spec.Procedure)
	buf.WriteString("\n  StreamType: ")
	buf.WriteString(spec.StreamType.String())
	buf.WriteByte('\n')
}

// logStreamEnd logs stream completion with duration and error if any.
func (l *Logger) logStreamEnd(buf *strings.Builder, err error, duration time.Duration) {
	buf.WriteString("[")
	buf.WriteString(time.Now().Format(l.timeFormat))
	buf.WriteString("] Stream completed in ")
	buf.WriteString(duration.String())
	buf.WriteByte('\n')

	if err != nil {
		l.logError(buf, err)
	}
}

// logProtoMessage formats protobuf messages with proper indentation.
func (l *Logger) logProtoMessage(buf *strings.Builder, title string, msg proto.Message) {
	if l.protoFormatter == nil {
		return
	}

	buf.WriteString("  ")
	buf.WriteString(title)
	buf.WriteString(":\n")

	data := l.protoFormatter.Format(msg)
	for line := range strings.SplitSeq(data, "\n") {
		if line != "" {
			buf.WriteString("    ")
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
}

// logHeaders logs HTTP headers with proper formatting.
func (l *Logger) logHeaders(buf *strings.Builder, title string, headers http.Header) {
	buf.WriteString("  ")
	buf.WriteString(title)
	buf.WriteString(":\n")

	for k, v := range headers {
		buf.WriteString("    ")
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(strings.Join(v, ", "))
		buf.WriteByte('\n')
	}
}

// logError logs detailed error information including Connect error metadata.
func (l *Logger) logError(buf *strings.Builder, err error) {
	buf.WriteString("  Error: ")
	buf.WriteString(err.Error())
	buf.WriteByte('\n')

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		buf.WriteString("  Code: ")
		buf.WriteString(connectErr.Code().String())
		buf.WriteByte('\n')

		if meta := connectErr.Meta(); len(meta) > 0 {
			buf.WriteString("  Metadata:\n")
			for k, v := range meta {
				buf.WriteString("    ")
				buf.WriteString(k)
				buf.WriteString(": ")
				buf.WriteString(strings.Join(v, ", "))
				buf.WriteByte('\n')
			}
		}
	}
}
