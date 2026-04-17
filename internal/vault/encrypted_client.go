package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// EncryptedClient wraps a SecretReader and transparently decrypts
// base64-encoded AES-GCM values before returning them to the caller.
type EncryptedClient struct {
	inner SecretReader
	gcm   cipher.AEAD
}

// NewEncryptedClient returns an EncryptedClient using the supplied 16, 24, or
// 32-byte AES key. Values that cannot be decoded/decrypted are left as-is so
// that plaintext fields are still accessible.
func NewEncryptedClient(inner SecretReader, key []byte) (*EncryptedClient, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypted_client: invalid key: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypted_client: gcm init: %w", err)
	}
	return &EncryptedClient{inner: inner, gcm: gcm}, nil
}

// ReadSecret fetches the secret then attempts to decrypt every string value.
func (c *EncryptedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := c.inner.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{}, len(data))
	for k, v := range data {
		if s, ok := v.(string); ok {
			if dec, err := c.decrypt(s); err == nil {
				out[k] = dec
				continue
			}
		}
		out[k] = v
	}
	return out, nil
}

func (c *EncryptedClient) decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	ns := c.gcm.NonceSize()
	if len(ciphertext) < ns {
		return "", fmt.Errorf("ciphertext too short")
	}
	plain, err := c.gcm.Open(nil, ciphertext[:ns], ciphertext[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
