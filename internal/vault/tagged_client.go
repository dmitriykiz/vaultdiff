package vault

import (
	"context"
	"fmt"
)

// TaggedClient wraps a SecretReader and injects a fixed set of string tags
// into every secret map returned under a reserved "_tags" key.
// Tags are expressed as key=value pairs and are appended as a single
// comma-separated string so downstream consumers can parse them easily.
type TaggedClient struct {
	inner SecretReader
	tags  map[string]string
}

// NewTaggedClient returns a TaggedClient that decorates inner with the
// provided tags. A nil or empty tags map is valid and results in no
// mutation of the returned secret data.
func NewTaggedClient(inner SecretReader, tags map[string]string) *TaggedClient {
	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
	}
	return &TaggedClient{inner: inner, tags: copy}
}

// ReadSecret delegates to the inner client and, on success, injects the
// tag string into the returned map under the "_tags" key. The original
// map is never mutated; a shallow copy is returned instead.
func (t *TaggedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := t.inner.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	if len(t.tags) == 0 {
		return data, nil
	}

	result := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		result[k] = v
	}
	result["_tags"] = buildTagString(t.tags)
	return result, nil
}

// Tags returns a copy of the tag map held by this client.
func (t *TaggedClient) Tags() map[string]string {
	copy := make(map[string]string, len(t.tags))
	for k, v := range t.tags {
		copy[k] = v
	}
	return copy
}

// buildTagString serialises tags into "k=v,k=v" form. Key order is
// non-deterministic; callers that need stable output should sort first.
func buildTagString(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	var out string
	for k, v := range tags {
		if out != "" {
			out += ","
		}
		out += fmt.Sprintf("%s=%s", k, v)
	}
	return out
}
