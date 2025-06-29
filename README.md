# devlog - ConnectRPC Call Logger

[![Go Reference](https://pkg.go.dev/badge/github.com/mdigger/connect-dev-log.svg)](https://pkg.go.dev/github.com/mdigger/connect-dev-log)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A feature-rich RPC call logger for ConnectRPC with protocol-agnostic support and structured logging capabilities.

## Features

- **Protocol Support**: Works with gRPC, Connect, and gRPC-Web protocols
- **Structured Logging**: Clear formatting for requests, responses, and errors
- **Protobuf Integration**: Built-in support for Protocol Buffer message formatting
- **Header/Metadata Logging**: Optional logging of request/response headers and metadata
- **Error Diagnostics**: Detailed error information including ConnectRPC error codes
- **Streaming Support**: Tracks bidirectional streaming messages with counters
- **Thread-Safe**: Safe for concurrent use across goroutines
- **Customizable**: Configurable timestamp formats, header visibility, and message formatting

## Installation

```bash
go get github.com/mdigger/connect-dev-log
```

## Usage

### Basic Setup

```go
import (
	"os"
	"connectrpc.com/connect"
	devlog "github.com/mdigger/connect-dev-log"
)

func main() {
	logger := devlog.New(os.Stdout)
	client := connect.NewClient[YourRequest, YourResponse](
		http.DefaultClient,
		"http://localhost:8080",
		connect.WithInterceptors(logger),
	)
	// Use client as normal...
}
```

### With Custom Options

```go
logger := devlog.New(
	os.Stdout,
	devlog.WithTimeFormat(time.DateTime),
	devlog.WithHeaders(true),
)
```

### Sample Output

```
[2023-11-15T14:23:45.123456Z] /package.Service/Method unary [::1]:54321
  Request:
    {"field":"value"}
  Response:
    {"result":"success"}
  Completed in: 12.345ms
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithTimeFormat` | Sets timestamp format (empty string disables) | `time.RFC3339Nano` |
| `WithHeaders` | Enables/disables header logging | `false` |
| `WithFormatter` | Custom protobuf message formatter | `protojson.MarshalOptions` |

## Streaming Support

The logger provides detailed tracking for streaming RPCs, including:
- Individual message logging for send/receive
- Message counters
- Stream duration metrics
- Client/Server disconnect detection

## Performance Considerations

- Uses a buffer pool to minimize allocations
- Thread-safe implementation with mutex protection

## Limitations

- Client-side streaming logging not yet implemented (`WrapStreamingClient`)
- No built-in log level filtering (handle at writer level)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
