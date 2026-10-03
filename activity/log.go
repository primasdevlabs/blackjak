package activity

import (
	"sync"
	"time"
)

// Entry is one activity log line for the activity window.
type Entry struct {
	ID        string    `json:"id"`
	RunID     string    `json:"runId,omitempty"`
	SessionID string    `json:"sessionId,omitempty"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// Ring is a fixed-size activity buffer.
type Ring struct {
	mu   sync.RWMutex
	max  int
	seq  int
	buf  []Entry
}

func NewRing(max int) *Ring {
	if max <= 0 {
		max = 1000
	}
	return &Ring{max: max, buf: make([]Entry, 0, max)}
}

func (r *Ring) Add(runID, typ, message string, data any) Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	e := Entry{
		ID:        formatID(r.seq),
		RunID:     runID,
		Type:      typ,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}
	if len(r.buf) >= r.max {
		r.buf = r.buf[1:]
	}
	r.buf = append(r.buf, e)
	return e
}

func (r *Ring) Tail(n int, runID string) []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if n <= 0 || n > len(r.buf) {
		n = len(r.buf)
	}
	src := r.buf
	if runID != "" {
		filtered := make([]Entry, 0, len(src))
		for _, e := range src {
			if e.RunID == runID {
				filtered = append(filtered, e)
			}
		}
		src = filtered
	}
	if n > len(src) {
		n = len(src)
	}
	out := make([]Entry, n)
	copy(out, src[len(src)-n:])
	return out
}

func formatID(n int) string {
	return time.Now().Format("150405") + "-" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
