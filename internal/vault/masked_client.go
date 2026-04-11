package vault

import (
	"context"
	"regexp"
	"strings"
)

// MaskRule defines a pattern and the replacement to apply when masking secret values.
type MaskRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// MaskedClient wraps a SecretReader and redacts sensitive values before returning them.
type MaskedClient struct {
	inner SecretReader
	rules []MaskRule
}

// NewMaskedClient creates a MaskedClient with the provided masking rules.
// If rules is nil or empty, values are returned unmodified.
func NewMaskedClient(inner SecretReader, rules []MaskRule) *MaskedClient {
	return &MaskedClient{inner: inner, rules: rules}
}

// ReadSecret fetches the secret and applies all mask rules to every string value.
func (m *MaskedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := m.inner.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	if len(m.rules) == 0 {
		return data, nil
	}
	masked := make(map[string]interface{}, len(data))
	for k, v := range data {
		if s, ok := v.(string); ok {
			masked[k] = m.applyRules(s)
		} else {
			masked[k] = v
		}
	}
	return masked, nil
}

func (m *MaskedClient) applyRules(value string) string {
	for _, rule := range m.rules {
		if rule.Pattern != nil {
			value = rule.Pattern.ReplaceAllString(value, rule.Replacement)
		} else {
			value = strings.ReplaceAll(value, value, rule.Replacement)
		}
	}
	return value
}
