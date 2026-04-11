package vault

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// AuditEvent represents a single audited vault operation.
type AuditEvent struct {
	Timestamp time.Time
	Operation string
	Path      string
	Err       error
}

func (e AuditEvent) String() string {
	status := "ok"
	if e.Err != nil {
		status = fmt.Sprintf("error: %v", e.Err)
	}
	return fmt.Sprintf("%s [%s] %s -> %s", e.Timestamp.Format(time.RFC3339), e.Operation, e.Path, status)
}

// AuditLog records vault operations for observability.
type AuditLog struct {
	mu     sync.Mutex
	events []AuditEvent
	w      io.Writer
}

// NewAuditLog creates an AuditLog that writes to w. Pass nil to suppress output.
func NewAuditLog(w io.Writer) *AuditLog {
	if w == nil {
		w = io.Discard
	}
	return &AuditLog{w: w}
}

// NewStderrAuditLog creates an AuditLog that writes to stderr.
func NewStderrAuditLog() *AuditLog {
	return NewAuditLog(os.Stderr)
}

// Record appends an event and writes it to the underlying writer.
func (a *AuditLog) Record(op, path string, err error) {
	e := AuditEvent{
		Timestamp: time.Now().UTC(),
		Operation: op,
		Path:      path,
		Err:       err,
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, e)
	fmt.Fprintln(a.w, e.String())
}

// Events returns a snapshot of all recorded events.
func (a *AuditLog) Events() []AuditEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditEvent, len(a.events))
	copy(out, a.events)
	return out
}

// Len returns the number of recorded events.
func (a *AuditLog) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.events)
}
