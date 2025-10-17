package devlog

import (
	"bytes"
	// "log/slog"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4096)) // Начальный capacity
	},
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

const maxBufferCapacity = 65536

func putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > maxBufferCapacity {
		// slog.Debug("rpc log buffer capacity overflow", slog.Int("cap", buf.Cap()))
		return
	}

	buf.Reset()
	bufferPool.Put(buf)
}
