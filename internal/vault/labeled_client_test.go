package vault

import (
	"errors"
	"testing"
)

// staticReader is a minimal SecretReader that returns a fixed map.
type staticReader struct {
	data map[string]interface{}
	err  error
}

func (s *staticReader) ReadSecret(_ string) (map[string]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	// Return a copy so mutations are detectable.
	out := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out, nil
}

func TestLabeledClient_NoLabels_ReturnsUnmodified(t *testing.T) {
	inner := &staticReader{data: map[string]interface{}{"key": "value"}}
	client := NewLabeledClient(inner, nil)

	got, err := client.ReadSecret("secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := got["_meta"]; ok {
		t.Error("expected no _meta key when labels are empty")
	}
	if got["key"] != "value" {
		t.Errorf("expected key=value, got %v", got["key"])
	}
}

func TestLabeledClient_InjectsMetaKey(t *testing.T) {
	inner := &staticReader{data: map[string]interface{}{"db_pass": "s3cr3t"}}
	client := NewLabeledClient(inner, map[string]string{"env": "staging", "region": "us-east-1"})

	got, err := client.ReadSecret("secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta, ok := got["_meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("_meta is missing or wrong type: %T", got["_meta"])
	}
	if meta["env"] != "staging" {
		t.Errorf("expected env=staging, got %v", meta["env"])
	}
	if meta["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %v", meta["region"])
	}
}

func TestLabeledClient_DoesNotMutateInnerResult(t *testing.T) {
	original := map[string]interface{}{"token": "abc"}
	inner := &staticReader{data: original}
	client := NewLabeledClient(inner, map[string]string{"source": "left"})

	got, _ := client.ReadSecret("secret/token")
	got["injected"] = true

	// Re-read from inner — should not contain the injected key.
	again, _ := inner.ReadSecret("secret/token")
	if _, found := again["injected"]; found {
		t.Error("LabeledClient mutated the inner client result")
	}
}

func TestLabeledClient_PropagatesError(t *testing.T) {
	sentinel := errors.New("vault unreachable")
	inner := &staticReader{err: sentinel}
	client := NewLabeledClient(inner, map[string]string{"env": "prod"})

	_, err := client.ReadSecret("secret/missing")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestLabeledClient_Labels_ReturnsCopy(t *testing.T) {
	client := NewLabeledClient(&staticReader{}, map[string]string{"a": "1"})
	labels := client.Labels()
	labels["a"] = "mutated"

	if client.Labels()["a"] != "1" {
		t.Error("Labels() returned a reference that allows mutation")
	}
}
