package vault

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// Logger is a simple structured logger interface.
type Logger interface {
	Log(level, msg string, fields map[string]any)
}

// StdLogger writes log lines to an io.Writer in a key=value format.
type StdLogger struct {
	w io.Writer
}

func NewStdLogger(w io.Writer) *StdLogger {
	if w == nil {
		w = os.Stderr
	}
	return &StdLogger{w: w}
}

func (l *StdLogger) Log(level, msg string, fields map[string]any) {
	line := fmt.Sprintf("time=%s level=%s msg=%q", time.Now().Format(time.RFC3339), level, msg)
	for k, v := range fields {
		line += fmt.Sprintf(" %s=%v", k, v)
	}
	fmt.Fprintln(l.w, line)
}

// LoggedClient wraps a SecretReader and logs every read operation.
type LoggedClient struct {
	inner  SecretReader
	logger Logger
}

func NewLoggedClient(inner SecretReader, logger Logger) *LoggedClient {
	if logger == nil {
		logger = NewStdLogger(io.Discard)
	}
	return &LoggedClient{inner: inner, logger: logger}
}

func (c *LoggedClient) ReadSecret(ctx context.Context, path string) (map[string]any, error) {
	start := time.Now()
	data, err := c.inner.ReadSecret(ctx, path)
	dur := time.Since(start)
	fields := map[string]any{
		"path":    path,
		"elapsed": dur.String(),
	}
	if err != nil {
		fields["error"] = err.Error()
		c.logger.Log("error", "vault read failed", fields)
		return nil, err
	}
	fields["keys"] = len(data)
	c.logger.Log("info", "vault read ok", fields)
	return data, nil
}
