package vault

import (
	"context"
	"fmt"
	"strings"
)

// EngineType represents the KV secrets engine version.
type EngineType int

const (
	EngineUnknown EngineType = iota
	EngineKVv1
	EngineKVv2
)

// DetectEngineType probes the Vault mount to determine whether it is a KV v1
// or KV v2 secrets engine. It returns EngineUnknown when the mount cannot be
// identified as either version.
func DetectEngineType(ctx context.Context, c *Client, mountPath string) (EngineType, error) {
	mountPath = strings.Trim(mountPath, "/")

	// KV v2 mounts expose a "metadata" sub-path.
	metaPath := fmt.Sprintf("%s/metadata", mountPath)
	_, err := c.logical.ReadWithContext(ctx, metaPath)
	if err == nil {
		return EngineKVv2, nil
	}

	// Fall back: try a direct read characteristic of KV v1.
	_, err = c.logical.ReadWithContext(ctx, mountPath)
	if err == nil {
		return EngineKVv1, nil
	}

	return EngineUnknown, fmt.Errorf("vault: unable to detect engine type for mount %q", mountPath)
}

// MountFromPath extracts the top-level mount segment from a full secret path.
// e.g. "secret/my/app" → "secret".
func MountFromPath(secretPath string) string {
	parts := strings.SplitN(strings.TrimPrefix(secretPath, "/"), "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
