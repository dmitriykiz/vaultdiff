package vault

import (
	"strings"
	"testing"
)

func TestEncryptor_RoundTrip(t *testing.T) {
	enc, err := NewEncryptor([]byte("abcdefghijklmnop"))
	if err != nil {
		t.Fatalf("NewEncryptor: %v", err)
	}
	original := "super-secret-value"
	ciphertext, err := enc.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ciphertext == original {
		t.Fatal("ciphertext should differ from plaintext")
	}
	plain, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if plain != original {
		t.Errorf("round-trip mismatch: got %q want %q", plain, original)
	}
}

func TestEncryptor_EncryptProducesUniqueNonces(t *testing.T) {
	enc, _ := NewEncryptor([]byte("abcdefghijklmnop"))
	a, _ := enc.Encrypt("hello")
	b, _ := enc.Encrypt("hello")
	if a == b {
		t.Error("two encryptions of same plaintext should differ due to random nonce")
	}
}

func TestEncryptor_DecryptInvalidBase64(t *testing.T) {
	enc, _ := NewEncryptor([]byte("abcdefghijklmnop"))
	_, err := enc.Decrypt("not-valid-base64!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestEncryptor_DecryptTamperedCiphertext(t *testing.T) {
	enc, _ := NewEncryptor([]byte("abcdefghijklmnop"))
	ct, _ := enc.Encrypt("value")
	// Flip last character to corrupt the ciphertext.
	tampered := ct[:len(ct)-1] + "X"
	if strings.HasSuffix(ct, "X") {
		tampered = ct[:len(ct)-1] + "Y"
	}
	_, err := enc.Decrypt(tampered)
	if err == nil {
		t.Fatal("expected authentication error for tampered ciphertext")
	}
}

func TestNewEncryptor_InvalidKey(t *testing.T) {
	_, err := NewEncryptor([]byte("tooshort"))
	if err == nil {
		t.Fatal("expected error for key that is too short")
	}
}
