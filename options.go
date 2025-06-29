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

// WithHeadersExcludes replaces the header value with `[** REDACTED **]`.
func WithHeaderExcludes(exclude []string) Option {
	ex := make(map[string]bool, len(exclude))
	for _, k := range exclude {
		ex[k] = true
	}

	return func(l *Logger) {
		l.excludeHeaders = ex
	}
}

// ProtoFormatter defines the interface for message formatting.
type ProtoFormatter interface {
	Format(proto.Message) string
}

// WithFormatter configures JSON, Text or other proto formatting settings.
func WithFormatter(opts ProtoFormatter) Option {
	return func(l *Logger) {
		l.protoFormatter = opts.Format
	}
}
