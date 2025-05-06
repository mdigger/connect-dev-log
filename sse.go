package devlog

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

var (
	pingMessage   = []byte(":\n\n")
	dataPrefix    = []byte("data: ")
	suffixNewline = []byte("\n\n")
)

type LogBroadcaster struct {
	clients      map[chan []byte]struct{}
	mu           sync.Mutex
	pingInterval time.Duration
}

func New(pingInterval time.Duration) *LogBroadcaster {
	if pingInterval <= 0 {
		pingInterval = time.Second * 15
	}
	return &LogBroadcaster{
		clients:      make(map[chan []byte]struct{}),
		pingInterval: pingInterval,
	}
}

func (lb *LogBroadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan []byte, 10)
	lb.addClient(clientChan)
	defer lb.removeClient(clientChan)

	pingTimer := time.NewTimer(lb.pingInterval)
	defer pingTimer.Stop()

	flusher := w.(http.Flusher)

	for {
		select {
		case msg := <-clientChan:
			w.Write(dataPrefix)
			w.Write(msg)
			w.Write(suffixNewline)
			flusher.Flush()

		case <-pingTimer.C:
			w.Write(pingMessage)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}

		pingTimer.Reset(lb.pingInterval)
	}
}

func (lb *LogBroadcaster) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.clients) == 0 {
		return len(p), nil
	}

	msg := bytes.Clone(p)

	for client := range lb.clients {
		select {
		case client <- msg:
		default:
			// skipping a message when the buffer overflows
		}
	}
	return len(p), nil
}

func (lb *LogBroadcaster) addClient(client chan []byte) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.clients[client] = struct{}{}
}

func (lb *LogBroadcaster) removeClient(client chan []byte) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	delete(lb.clients, client)
	close(client)
}
