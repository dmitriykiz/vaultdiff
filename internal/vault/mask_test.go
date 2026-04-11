package vault

import (
	"testing"
)

func TestSensitiveKeyMatcher_DetectsPassword(t *testing.T) {
	keys := []string{"password", "db_password", "PASSWORD", "UserPassword"}
	for _, k := range keys {
		if !SensitiveKeyMatcher(k) {
			t.Errorf("expected %q to be sensitive", k)
		}
	}
}

func TestSensitiveKeyMatcher_AllowsNonSensitive(t *testing.T) {
	keys := []string{"username", "host", "port", "database", "region"}
	for _, k := range keys {
		if SensitiveKeyMatcher(k) {
			t.Errorf("expected %q to be non-sensitive", k)
		}
	}
}

func TestRedactSensitiveKeys_RedactsMatchingKeys(t *testing.T) {
	data := map[string]interface{}{
		"password": "hunter2",
		"username": "alice",
		"api_key":  "abc123",
		"host":     "localhost",
	}
	out := RedactSensitiveKeys(data, "[REDACTED]")
	if out["password"] != "[REDACTED]" {
		t.Errorf("expected password redacted, got %v", out["password"])
	}
	if out["api_key"] != "[REDACTED]" {
		t.Errorf("expected api_key redacted, got %v", out["api_key"])
	}
	if out["username"] != "alice" {
		t.Errorf("expected username unchanged, got %v", out["username"])
	}
	if out["host"] != "localhost" {
		t.Errorf("expected host unchanged, got %v", out["host"])
	}
}

func TestRedactSensitiveKeys_DoesNotMutateOriginal(t *testing.T) {
	data := map[string]interface{}{"password": "secret"}
	_ = RedactSensitiveKeys(data, "[REDACTED]")
	if data["password"] != "secret" {
		t.Error("original map was mutated")
	}
}

func TestCommonMaskRules_MasksJWT(t *testing.T) {
	rules := CommonMaskRules()
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	masked := &MaskedClient{rules: rules}
	result := masked.applyRules(token)
	if result == token {
		t.Error("expected JWT to be masked")
	}
}

func TestCommonMaskRules_MasksHexSecret(t *testing.T) {
	rules := CommonMaskRules()
	hex := "a3f1c2d4e5b6789012345678abcdef01"
	masked := &MaskedClient{rules: rules}
	result := masked.applyRules(hex)
	if result == hex {
		t.Error("expected hex secret to be masked")
	}
}
