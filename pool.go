package devlog

import (
	"strings"
	"sync"
)

var builderPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}

// getBuilder gets a strings.Builder from the shared pool.
// The returned builder should be returned to the pool using putBuilder.
func getBuilder() *strings.Builder {
	return builderPool.Get().(*strings.Builder) //nolint:forcetypeassert,errcheck
}

// putBuilder returns a strings.Builder to the shared pool after resetting it.
// Always call this function when done with a builder obtained from getBuilder.
func putBuilder(b *strings.Builder) {
	b.Reset()
	builderPool.Put(b)
}
