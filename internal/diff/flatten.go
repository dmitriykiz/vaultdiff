package diff

import "fmt"

// FlattenSecrets converts a nested map[string]map[string]interface{} (path →
// kv-data) into a flat map keyed by "<path>.<field>" strings.  This makes it
// easy to feed the result directly into Compare.
func FlattenSecrets(secrets map[string]map[string]interface{}) map[string]string {
	out := make(map[string]string, len(secrets))
	for path, data := range secrets {
		for key, val := range data {
			flat := fmt.Sprintf("%s.%s", path, key)
			out[flat] = stringify(val)
		}
	}
	return out
}

// stringify converts an arbitrary secret value to a stable string
// representation suitable for diffing.
func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		// Vault JSON numbers arrive as float64; trim unnecessary decimals.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
