package vault

import (
	"regexp"
	"strings"
)

// SanitizeRule defines a rule for sanitizing secret values.
type SanitizeRule struct {
	// KeyPattern is a regex matched against the secret key.
	KeyPattern *regexp.Regexp
	// Replacement is the string used to replace matched values.
	Replacement string
}

// Sanitizer applies a set of sanitize rules to secret data.
type Sanitizer struct {
	rules []SanitizeRule
}

// NewSanitizer creates a Sanitizer with the provided rules.
func NewSanitizer(rules []SanitizeRule) *Sanitizer {
	return &Sanitizer{rules: rules}
}

// Apply returns a copy of data with values replaced according to rules.
func (s *Sanitizer) Apply(data map[string]interface{}) map[string]interface{} {
	if s == nil || len(s.rules) == 0 {
		return data
	}
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		out[k] = s.sanitizeValue(k, v)
	}
	return out
}

func (s *Sanitizer) sanitizeValue(key string, value interface{}) interface{} {
	for _, rule := range s.rules {
		if rule.KeyPattern != nil && rule.KeyPattern.MatchString(strings.ToLower(key)) {
			return rule.Replacement
		}
	}
	return value
}

// DefaultSanitizeRules returns a standard set of sanitization rules.
func DefaultSanitizeRules() []SanitizeRule {
	return []SanitizeRule{
		{
			KeyPattern:  regexp.MustCompile(`(password|passwd|secret|token|api_?key|private_?key|credential)`),
			Replacement: "[sanitized]",
		},
	}
}
