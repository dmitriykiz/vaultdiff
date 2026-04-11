package vault

import (
	"regexp"
	"strings"
)

// CommonMaskRules returns a default set of MaskRules that redact values for
// keys commonly associated with sensitive data (passwords, tokens, keys, etc.).
func CommonMaskRules() []MaskRule {
	return []MaskRule{
		{
			// Mask anything that looks like a JWT (three base64 segments)
			Pattern:     regexp.MustCompile(`[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),
			Replacement: "[JWT_REDACTED]",
		},
		{
			// Mask 32+ hex character strings (API keys, secrets)
			Pattern:     regexp.MustCompile(`[0-9a-fA-F]{32,}`),
			Replacement: "[HEX_REDACTED]",
		},
	}
}

// SensitiveKeyMatcher returns true if the given key name suggests it holds
// a sensitive value (password, token, secret, key, etc.).
func SensitiveKeyMatcher(key string) bool {
	lower := strings.ToLower(key)
	sensitiveTerms := []string{
		"password", "passwd", "secret", "token",
		"apikey", "api_key", "private_key", "privatekey",
		"credential", "auth",
	}
	for _, term := range sensitiveTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

// RedactSensitiveKeys returns a copy of data where values for sensitive keys
// are replaced with the provided placeholder.
func RedactSensitiveKeys(data map[string]interface{}, placeholder string) map[string]interface{} {
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		if SensitiveKeyMatcher(k) {
			out[k] = placeholder
		} else {
			out[k] = v
		}
	}
	return out
}
