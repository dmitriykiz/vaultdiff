package vault

import (
	"context"
)

// AuditedClient wraps a SecretReader and records every operation to an AuditLog.
type AuditedClient struct {
	inner SecretReader
	log   *AuditLog
}

// NewAuditedClient wraps inner with audit logging.
func NewAuditedClient(inner SecretReader, log *AuditLog) *AuditedClient {
	if log == nil {
		log = NewAuditLog(nil)
	}
	return &AuditedClient{inner: inner, log: log}
}

// ReadSecret delegates to the inner client and records the result.
func (a *AuditedClient) ReadSecret(ctx context.Context, path string) (map[string]string, error) {
	data, err := a.inner.ReadSecret(ctx, path)
	a.log.Record("read", path, err)
	return data, err
}

// Log returns the underlying AuditLog for inspection.
func (a *AuditedClient) Log() *AuditLog {
	return a.log
}
