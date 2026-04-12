package vault

// LabeledClient wraps a SecretReader and attaches a fixed set of metadata
// labels to every secret returned. Labels are merged into the secret map
// under a reserved "_meta" key so downstream consumers (diff, formatters)
// can identify the source of each value without changing the primary data.
type LabeledClient struct {
	inner  SecretReader
	labels map[string]string
}

// NewLabeledClient creates a LabeledClient that decorates every ReadSecret
// response with the provided labels. A nil or empty labels map is valid and
// results in a no-op decoration.
func NewLabeledClient(inner SecretReader, labels map[string]string) *LabeledClient {
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return &LabeledClient{inner: inner, labels: copy}
}

// ReadSecret delegates to the inner client and then injects the label map
// into the returned data under the "_meta" key. If the inner client returns
// an error the error is propagated unchanged.
func (lc *LabeledClient) ReadSecret(path string) (map[string]interface{}, error) {
	data, err := lc.inner.ReadSecret(path)
	if err != nil {
		return nil, err
	}

	if len(lc.labels) == 0 {
		return data, nil
	}

	// Shallow-copy the returned map so we do not mutate the inner client's
	// cached value.
	out := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		out[k] = v
	}

	meta := make(map[string]interface{}, len(lc.labels))
	for k, v := range lc.labels {
		meta[k] = v
	}
	out["_meta"] = meta

	return out, nil
}

// Labels returns a copy of the label set attached to this client.
func (lc *LabeledClient) Labels() map[string]string {
	copy := make(map[string]string, len(lc.labels))
	for k, v := range lc.labels {
		copy[k] = v
	}
	return copy
}
