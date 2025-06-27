ConnectRPC Dev Logger
=====================

[![Go Reference](https://pkg.go.dev/badge/github.com/mdigger/connect-dev-log.svg)](https://pkg.go.dev/github.com/mdigger/connect-dev-log)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A high-performance logging interceptor for ConnectRPC with protocol-agnostic request/response logging, protobuf message formatting, and streaming support.

## Features

- **Full ConnectRPC Interceptor Implementation**
  - Supports both unary and streaming RPCs
  - Works with gRPC, Connect, and gRPC-Web protocols

- **Detailed Request/Response Logging**
  - Structured logging with timestamps
  - Protocol and stream type information
  - Peer address and HTTP method logging

- **Protobuf Message Formatting**
  - Supports both Text and JSON formats
  - Configurable indentation and formatting
  - Automatic error handling

- **Efficient Design**
  - Zero-allocation string building
  - Thread-safe with sync.Pool for buffers
  - Minimal performance overhead

- **Flexible Configuration**
  - Customizable timestamp format
  - Optional header/metadata logging
  - Extensible formatter interface

## Installation

```bash
go get github.com/mdigger/connect-dev-log
```

## Usage

### Basic Setup

```go
import (
	"os"
	"github.com/mdigger/connect-dev-log"
)

// Create logger with default settings (json format)
logger := devlog.New(os.Stdout)

// Use with Connect client
client := pingv1connect.NewPingServiceClient(
	http.DefaultClient,
	"http://localhost:8080",
	connect.WithInterceptors(logger),
)

// Or with handler
mux := http.NewServeMux()
mux.Handle(pingv1connect.NewPingServiceHandler(
	&pingServer{},
	connect.WithInterceptors(logger),
))
```

### Advanced Configuration

```go
// JSON format with custom options
logger := devlog.New(os.Stdout,
	devlog.WithHeaders(true),
	devlog.WithTimeFormat(time.DateTime),
	devlog.WithFormatter(protojson.MarshalOptions{
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}),
)

// Or with text format
logger := devlog.New(os.Stdout,
	devlog.WithFormatter(prototext.MarshalOptions{
		Indent:    "    ",
		Multiline: true,
	}),
)
```

### Custom Formatters

Implement the ProtoFormatter interface for custom formatting:

```go
type CompactFormatter struct{}

func (f CompactFormatter) Format(m proto.Message) string {
	return proto.CompactTextString(m)
}

// Usage:
logger := devlog.New(os.Stdout,
	devlog.WithFormatter(CompactFormatter{}),
)
```

## Example Output

### Text Format

```
[2025-07-27T14:23:45.123Z] /package.Service/Method from 127.0.0.1:12345
  StreamType: Unary
  HTTPMethod: POST
  Request message:
    field1: "value"
    field2: 42
```

### JSON Format

```
[2023-11-15T14:23:45.123Z] /package.Service/Method from 127.0.0.1:12345
  StreamType: Unary
  HTTPMethod: POST
  Request message:
    {
      "field1": "value",
      "field2": 42
    }
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
