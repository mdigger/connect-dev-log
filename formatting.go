package devlog

import (
	"bytes"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

func (l *Logger) writeTimestamp(w *bytes.Buffer, time time.Time) {
	if l.timeFormat == "" {
		return
	}
	w.WriteByte('[')
	w.WriteString(time.Format(l.timeFormat))
	w.WriteString("] ")
}

func (l *Logger) writeHeaders(w *bytes.Buffer, headers http.Header, exclude map[string]bool) {
	w.WriteString("\n  Headers:")
	keys := slices.Sorted(maps.Keys(headers))
	for _, k := range keys {
		w.WriteString("\n    ")
		w.WriteString(k)
		w.WriteString(": ")
		if exclude[k] {
			w.WriteString("[** redacted **]")
			continue
		}
		w.WriteString(strings.Join(headers.Values(k), ", "))
	}
}

func (l *Logger) writeProto(w *bytes.Buffer, proto proto.Message) {
	message := l.protoFormatter(proto)
	for line := range strings.SplitSeq(message, "\n") {
		if line != "" {
			w.WriteString("\n    ")
			w.WriteString(line)
		} else {
			slog.Warn("rpc <empty proto line>", slog.String("proto", message))
		}
	}
}

func (l *Logger) writeError(w *bytes.Buffer, err error) {
	w.WriteString("\n  Error: ")
	w.WriteString(err.Error())

	// var connectErr *connect.Error
	// if errors.As(err, &connectErr) {
	// 	w.WriteString("\n  Code: ")
	// 	w.WriteString(connectErr.Code().String())
	// 	w.WriteString("\n  Message: ")
	// 	w.WriteString(connectErr.Message())

	// 	if meta := connectErr.Meta(); len(meta) > 0 {
	// 		l.writeHeaders(w, meta, nil)
	// 	}

	// 	if details := connectErr.Details(); len(details) > 0 {
	// 		for _, detail := range details {
	// 			w.WriteString("\n    Type: ")
	// 			w.WriteString(detail.Type())
	// 			v, err := detail.Value()
	// 			if err != nil {
	// 				slog.Warn("error getting value from detail",
	// 					slog.String("type", detail.Type()),
	// 					slog.Any("error", err))
	// 			} else {
	// 				w.WriteString("\n    Message: ")
	// 				l.writeProto(w, v)
	// 			}
	// 		}
	// 	}
	// }
}
