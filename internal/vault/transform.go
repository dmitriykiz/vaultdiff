package vault

import "strings"

// TransformRule defines a key-based transformation applied to secret values.
type TransformRule struct {
	// KeySuffix is matched against the end of each key (case-insensitive).
	KeySuffix string
	// Transform is applied to the string representation of the value.
	Transform func(string) string
}

// Transformer applies a set of TransformRules to secret data maps.
type Transformer struct {
	rules []TransformRule
}

// NewTransformer returns a Transformer that applies the given rules in order.
func NewTransformer(rules []TransformRule) *Transformer {
	return &Transformer{rules: rules}
}

// Apply returns a new map with transformation rules applied to matching keys.
// Non-string values are passed through unchanged.
func (t *Transformer) Apply(data map[string]interface{}) map[string]interface{} {
	if t == nil || len(t.rules) == 0 {
		return data
	}
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		out[k] = t.applyKey(k, v)
	}
	return out
}

func (t *Transformer) applyKey(key string, value interface{}) interface{} {
	s, ok := value.(string)
	if !ok {
		return value
	}
	lower := strings.ToLower(key)
	for _, r := range t.rules {
		if strings.HasSuffix(lower, strings.ToLower(r.KeySuffix)) {
			s = r.Transform(s)
		}
	}
	return s
}

// TruncateTransform returns a TransformRule that truncates values longer than n characters.
func TruncateTransform(suffix string, n int) TransformRule {
	return TransformRule{
		KeySuffix: suffix,
		Transform: func(s string) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "..."
		},
	}
}

// UpperCaseTransform returns a TransformRule that upper-cases matching values.
func UpperCaseTransform(suffix string) TransformRule {
	return TransformRule{
		KeySuffix: suffix,
		Transform: strings.ToUpper,
	}
}
