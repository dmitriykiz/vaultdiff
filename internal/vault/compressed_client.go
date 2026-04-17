package vault

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
)

// CompressedClient wraps a SecretReader and compresses/decompresses string
// values in the returned secret map using gzip + base64 encoding.
type CompressedClient struct {
	inner  SecretReader
	enabled bool
}

// NewCompressedClient returns a CompressedClient wrapping inner.
// When enabled is false the client is a transparent pass-through.
func NewCompressedClient(inner SecretReader, enabled bool) *CompressedClient {
	return &CompressedClient{inner: inner, enabled: enabled}
}

func (c *CompressedClient) ReadSecret(ctx context.Context, path string) (map[string]any, error) {
	data, err := c.inner.ReadSecret(ctx, path)
	if err != nil || !c.enabled {
		return data, err
	}
	return decompressValues(data)
}

func decompressValues(data map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(data))
	for k, v := range data {
		str, ok := v.(string)
		if !ok {
			out[k] = v
			continue
		}
		decoded, err := tryGunzip(str)
		if err != nil {
			// not compressed – keep original
			out[k] = v
			continue
		}
		out[k] = decoded
	}
	return out, nil
}

func tryGunzip(s string) (any, error) {
	r, err := gzip.NewReader(bytes.NewBufferString(s))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		return buf.String(), nil
	}
	return v, nil
}

func compressString(s string) (string, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := fmt.Fprint(w, s); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
