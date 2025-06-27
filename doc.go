// Package devlog provides a high-performance RPC call logger for ConnectRPC.
//
// It implements connect.Interceptor for both unary and streaming RPCs with:
// - Detailed protocol-agnostic logging (gRPC, Connect, gRPC-Web)
// - Structured request/response logging
// - Protobuf message formatting
// - Header and metadata support
// - Error diagnostics
// - Streaming message tracking
package devlog
