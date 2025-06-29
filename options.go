package devlog

import "google.golang.org/protobuf/proto"

// Option configures Logger behavior.
type Option func(*Logger)

// WithTimeFormat sets the timestamp format (default: RFC3339Nano).
func WithTimeFormat(format string) Option {
	return func(l *Logger) {
		l.timeFormat = format
	}
}

// WithHeaders enables/disables header logging (default: false).
func WithHeaders(show bool) Option {
	return func(l *Logger) {
		l.showHeaders = show
	}
}

// ProtoFormatter defines the interface for message formatting.
type ProtoFormatter interface {
	Format(proto.Message) string
}

// WithFormatter configures JSON, Text or other proto formatting settings.
func WithFormatter(opts ProtoFormatter) Option {
	return func(l *Logger) {
		l.protoFormatter = opts
	}
}
