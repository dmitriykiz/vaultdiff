package vault

import (
	"context"
	"errors"
	"strings"
)

// ValidationRule is a function that validates a secret path or its data.
type ValidationRule func(path string, data map[string]interface{}) error

// ValidatedClient wraps a SecretReader and applies validation rules after
// each successful read. If any rule returns an error the read is failed.
type ValidatedClient struct {
	inner SecretReader
	rules []ValidationRule
}

// NewValidatedClient returns a ValidatedClient that applies the given rules
// to every secret returned by inner.
func NewValidatedClient(inner SecretReader, rules ...ValidationRule) *ValidatedClient {
	return &ValidatedClient{inner: inner, rules: rules}
}

// ReadSecret reads from the inner client and validates the result.
func (v *ValidatedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := v.inner.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	for _, rule := range v.rules {
		if rErr := rule(path, data); rErr != nil {
			return nil, rErr
		}
	}
	return data, nil
}

// RequireKeys returns a ValidationRule that errors when any of the given keys
// are absent from the secret data.
func RequireKeys(keys ...string) ValidationRule {
	return func(path string, data map[string]interface{}) error {
		var missing []string
		for _, k := range keys {
			if _, ok := data[k]; !ok {
				missing = append(missing, k)
			}
		}
		if len(missing) > 0 {
			return errors.New("secret at " + path + " missing required keys: " + strings.Join(missing, ", "))
		}
		return nil
	}
}

// DenyEmptyValues returns a ValidationRule that errors when any value in the
// secret data is an empty string.
func DenyEmptyValues() ValidationRule {
	return func(path string, data map[string]interface{}) error {
		for k, v := range data {
			if s, ok := v.(string); ok && s == "" {
				return errors.New("secret at " + path + " has empty value for key: " + k)
			}
		}
		return nil
	}
}
