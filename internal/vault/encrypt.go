package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor provides AES-GCM encryption helpers used by EncryptedClient and
// any tooling that needs to pre-encrypt values before writing to Vault.
type Encryptor struct {
	gcm cipher.AEAD
}

// NewEncryptor constructs an Encryptor from a 16, 24, or 32-byte AES key.
func NewEncryptor(key []byte) (*Encryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encryptor: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encryptor: gcm: %w", err)
	}
	return &Encryptor{gcm: gcm}, nil
}

// Encrypt seals plaintext with a random nonce and returns a base64-encoded
// ciphertext string suitable for storage in Vault.
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("encryptor: nonce: %w", err)
	}
	ct := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt reverses Encrypt. Returns an error if the ciphertext is invalid or
// authentication fails.
func (e *Encryptor) Decrypt(encoded string) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("encryptor: base64: %w", err)
	}
	ns := e.gcm.NonceSize()
	if len(ct) < ns {
		return "", fmt.Errorf("encryptor: ciphertext too short")
	}
	plain, err := e.gcm.Open(nil, ct[:ns], ct[ns:], nil)
	if err != nil {
		return "", fmt.Errorf("encryptor: open: %w", err)
	}
	return string(plain), nil
}
