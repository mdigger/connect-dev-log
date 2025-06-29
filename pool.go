package devlog

import (
	"bytes"
	"log/slog"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 256)) // Начальный capacity
	},
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 4096 {
		slog.Warn("buffer capacity", slog.Int("cap", buf.Cap()))
		buf = nil
	} else {
		buf.Reset()
	}

	bufferPool.Put(buf)
}
