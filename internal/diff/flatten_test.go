package diff

import (
	"testing"
)

func TestFlattenSecrets_BasicKeys(t *testing.T) {
	input := map[string]map[string]interface{}{
		"secret/app": {
			"db_pass": "hunter2",
			"port":    float64(5432),
		},
	}

	got := FlattenSecrets(input)

	if v, ok := got["secret/app.db_pass"]; !ok || v != "hunter2" {
		t.Errorf("expected secret/app.db_pass=hunter2, got %q", v)
	}
	if v, ok := got["secret/app.port"]; !ok || v != "5432" {
		t.Errorf("expected secret/app.port=5432, got %q", v)
	}
}

func TestFlattenSecrets_MultipleSecrets(t *testing.T) {
	input := map[string]map[string]interface{}{
		"secret/a": {"x": "1"},
		"secret/b": {"y": "2"},
	}
	got := FlattenSecrets(input)
	if len(got) != 2 {
		t.Errorf("expected 2 keys, got %d", len(got))
	}
}

func TestFlattenSecrets_EmptyInput(t *testing.T) {
	got := FlattenSecrets(map[string]map[string]interface{}{})
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestStringify_Types(t *testing.T) {
	cases := []struct {
		input interface{}
		want  string
	}{
		{nil, ""},
		{"hello", "hello"},
		{true, "true"},
		{false, "false"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{[]string{"a", "b"}, "[a b]"},
	}
	for _, tc := range cases {
		got := stringify(tc.input)
		if got != tc.want {
			t.Errorf("stringify(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFlattenSecrets_NilValue(t *testing.T) {
	input := map[string]map[string]interface{}{
		"secret/cfg": {"empty": nil},
	}
	got := FlattenSecrets(input)
	if v, ok := got["secret/cfg.empty"]; !ok || v != "" {
		t.Errorf("expected empty string for nil value, got %q", v)
	}
}
