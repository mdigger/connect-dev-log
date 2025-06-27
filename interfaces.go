package devlog

import "connectrpc.com/connect"

// Verify Logger fully implements connect.Interceptor.
var _ connect.Interceptor = (*Logger)(nil)
