package devlog

import (
	"bytes"
	"cmp"
	"context"
	"io"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
)

const (
	headerWidth   = 60
	initialBufCap = 1024
)

type logBuilder struct {
	buf    *bytes.Buffer
	tw     *tabwriter.Writer
	config *InterceptorConfig
}

var bytePool = sync.Pool{
	New: func() any {
		return make([]byte, 0, initialBufCap)
	},
}

func getBuilder(config *InterceptorConfig) *logBuilder {
	b := bytePool.Get().([]byte)
	b = b[:0]

	buf := bytes.NewBuffer(b)
	return &logBuilder{
		buf:    buf,
		tw:     tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0),
		config: config,
	}
}

func putBuilder(b *logBuilder) {
	if b.config.MaxBufferSize == 0 || cap(b.buf.Bytes()) <= b.config.MaxBufferSize {
		bytePool.Put(b.buf.Bytes())
	}
}

func (lb *logBuilder) writeHeaders(headers map[string][]string) {
	if len(headers) == 0 {
		return
	}

	lb.writeSubheader("Headers")

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}

	// Сортируем (регистронезависимо)
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	for _, k := range keys {
		value := strings.Join(headers[k], ", ")
		if slices.Contains(lb.config.HiddenHeaders, strings.ToLower(k)) {
			value = lb.config.RedactValue(value)
		}
		lb.writeKeyValue(k, value)
	}
	lb.flush()
}

func (lb *logBuilder) writeContextData(ctx context.Context, extractor func(context.Context) map[string]string) {
	if extractor == nil {
		return
	}

	data := extractor(ctx)
	if len(data) == 0 {
		return
	}

	lb.writeSubheader("Context Data")

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		lb.writeKeyValue(k, data[k])
	}
	lb.flush()
}

// Остальные методы остаются без изменений
func (lb *logBuilder) writeMainHeader(title string) {
	padding := max(0, headerWidth-len(title)-6)
	lb.buf.WriteString("== " + title + " " + strings.Repeat("=", padding) + "==\n")
}

func (lb *logBuilder) writeSubheader(title string) {
	padding := max(0, headerWidth-len(title)-6)
	lb.buf.WriteString("-- " + title + " " + strings.Repeat("-", padding) + "--\n")
}

func (lb *logBuilder) writeKeyValue(key, value string) {
	lb.tw.Write([]byte(key + ":\t" + value + "\n"))
}

func (lb *logBuilder) writeRaw(data string) {
	lb.buf.WriteString(data + "\n")
}

func (lb *logBuilder) flush() {
	lb.tw.Flush()
}

func (lb *logBuilder) writeTo(w io.Writer) {
	lb.flush()
	w.Write(lb.buf.Bytes())
}
