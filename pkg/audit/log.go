package audit

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/akram/kgb/pkg/intent"
)

// Entry records a single evaluation outcome.
type Entry struct {
	Timestamp time.Time       `json:"ts"`
	Namespace string          `json:"namespace"`
	Session   string          `json:"session,omitempty"`
	Intent    intent.Intent   `json:"intent"`
	Outcome   string          `json:"outcome"`
	Policy    string          `json:"policy,omitempty"`
	Reason    string          `json:"reason,omitempty"`
	Replay    bool            `json:"replay,omitempty"`
}

// Logger writes JSON lines to stdout (ship to Loki/ELK via cluster logging).
type Logger struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// NewLogger builds an audit logger.
func NewLogger() *Logger {
	return &Logger{enc: json.NewEncoder(os.Stdout)}
}

// Write emits one JSON line.
func (l *Logger) Write(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	if err := l.enc.Encode(e); err != nil {
		slog.Error("audit encode failed", "error", err)
	}
}
