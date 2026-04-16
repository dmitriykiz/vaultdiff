package vault

import (
	"context"
	"fmt"
	"io"
	"time"
)

// TraceEvent records a single read operation for tracing purposes.
type TraceEvent struct {
	Path     string
	Duration time.Duration
	Err      error
}

// Tracer collects trace events emitted by TracedClient.
type Tracer struct {
	w      io.Writer
	events []TraceEvent
}

// NewTracer creates a Tracer that writes human-readable trace lines to w.
// Pass nil to collect events silently.
func NewTracer(w io.Writer) *Tracer {
	return &Tracer{w: w}
}

// record appends an event and optionally writes it.
func (t *Tracer) record(e TraceEvent) {
	t.events = append(t.events, e)
	if t.w == nil {
		return
	}
	status := "ok"
	if e.Err != nil {
		status = fmt.Sprintf("err: %v", e.Err)
	}
	fmt.Fprintf(t.w, "[trace] path=%s duration=%s status=%s\n", e.Path, e.Duration.Round(time.Microsecond), status)
}

// Events returns a copy of all collected trace events.
func (t *Tracer) Events() []TraceEvent {
	out := make([]TraceEvent, len(t.events))
	copy(out, t.events)
	return out
}

// TracedClient wraps a SecretReader and records timing for every read.
type TracedClient struct {
	inner  SecretReader
	tracer *Tracer
}

// NewTracedClient returns a TracedClient. If tracer is nil a no-op tracer is used.
func NewTracedClient(inner SecretReader, tracer *Tracer) *TracedClient {
	if tracer == nil {
		tracer = NewTracer(nil)
	}
	return &TracedClient{inner: inner, tracer: tracer}
}

// ReadSecret delegates to the inner client and records a TraceEvent.
func (c *TracedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	start := time.Now()
	data, err := c.inner.ReadSecret(ctx, path)
	c.tracer.record(TraceEvent{
		Path:     path,
		Duration: time.Since(start),
		Err:      err,
	})
	return data, err
}

// Tracer exposes the underlying Tracer for inspection.
func (c *TracedClient) Tracer() *Tracer { return c.tracer }
