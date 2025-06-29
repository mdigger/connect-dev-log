package devlog

import (
	"bytes"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"

	"connectrpc.com/connect"
)

func writeTimestamp(w *bytes.Buffer, timestamp string) {
	w.WriteByte('[')
	w.WriteString(timestamp)
	w.WriteString("] ")
}

func writeHeaders(w *bytes.Buffer, headers http.Header, exclude map[string]bool) {
	w.WriteString("\n  Headers:")
	keys := slices.Sorted(maps.Keys(headers))
	for _, k := range keys {
		w.WriteString("\n    ")
		w.WriteString(k)
		w.WriteString(": ")
		if exclude[k] {
			w.WriteString("[** REDACTED **]")
			continue
		}
		w.WriteString(strings.Join(headers.Values(k), ", "))
	}
}

func writeProto(w *bytes.Buffer, proto string) {
	for line := range strings.SplitSeq(proto, "\n") {
		if line != "" {
			w.WriteString("\n    ")
			w.WriteString(line)
		} else {
			slog.Warn("empty proto line", slog.String("proto", proto))
		}
	}
}

func writeError(w *bytes.Buffer, err error) {
	w.WriteString("\n  Error: ")
	w.WriteString(err.Error())

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		w.WriteString("\n  Code: ")
		w.WriteString(connectErr.Code().String())

		if meta := connectErr.Meta(); len(meta) > 0 {
			w.WriteString("\n  Metadata:")
			for k, v := range meta {
				w.WriteString("\n    ")
				w.WriteString(k)
				w.WriteString(": ")
				w.WriteString(strings.Join(v, ", "))
			}
		}
	}
}
