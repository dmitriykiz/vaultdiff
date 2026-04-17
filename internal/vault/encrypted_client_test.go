package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func encryptValue(t *testing.T, key []byte, plain string) string {
	t.Helper()
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)
	ct := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ct)
}

func newEncryptTestServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeKVv1Response(w, data)
	}))
}

func TestEncryptedClient_DecryptsStringValues(t *testing.T) {
	key := []byte("0123456789abcdef") // 16-byte AES-128
	enc := encryptValue(t, key, "s3cr3t")
	srv := newEncryptTestServer(map[string]interface{}{"password": enc, "count": 42})
	defer srv.Close()

	base := newTestClient(t, srv.URL)
	client, err := NewEncryptedClient(base, key)
	if err != nil {
		t.Fatalf("NewEncryptedClient: %v", err)
	}
	data, err := client.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("ReadSecret: %v", err)
	}
	if data["password"] != "s3cr3t" {
		t.Errorf("expected decrypted value, got %v", data["password"])
	}
	if data["count"] != 42 {
		t.Errorf("non-string value mutated: %v", data["count"])
	}
}

func TestEncryptedClient_PlaintextLeftUnchanged(t *testing.T) {
	key := []byte("0123456789abcdef")
	srv := newEncryptTestServer(map[string]interface{}{"token": "plaintext-value"})
	defer srv.Close()

	base := newTestClient(t, srv.URL)
	client, _ := NewEncryptedClient(base, key)
	data, err := client.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("ReadSecret: %v", err)
	}
	if data["token"] != "plaintext-value" {
		t.Errorf("expected original value, got %v", data["token"])
	}
}

func TestEncryptedClient_InvalidKey(t *testing.T) {
	_, err := NewEncryptedClient(nil, []byte("short"))
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestEncryptedClient_PropagatesInnerError(t *testing.T) {
	key := []byte("0123456789abcdef")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	base := newTestClient(t, srv.URL)
	client, _ := NewEncryptedClient(base, key)
	_, err := client.ReadSecret(context.Background(), "secret/forbidden")
	if err == nil {
		t.Fatal("expected error from inner client")
	}
}
