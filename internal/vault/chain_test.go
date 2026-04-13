package vault

import (
	"context"
	"errors"
	"testing"
)

func TestChainClient_FirstReaderSucceeds(t *testing.T) {
	r1 := &staticDataReader{data: map[string]interface{}{"src": "first"}}
	r2 := &staticDataReader{data: map[string]interface{}{"src": "second"}}

	c, err := NewChainClient(r1, r2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := c.ReadSecret(context.Background(), "secret/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["src"] != "first" {
		t.Errorf("expected first reader, got %v", data)
	}
}

func TestChainClient_SkipsFailedReaders(t *testing.T) {
	sentinel := errors.New("unavailable")
	r1 := &staticErrorReader{err: sentinel}
	r2 := &staticDataReader{data: map[string]interface{}{"src": "second"}}

	c, _ := NewChainClient(r1, r2)
	data, err := c.ReadSecret(context.Background(), "secret/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["src"] != "second" {
		t.Errorf("expected second reader, got %v", data)
	}
}

func TestChainClient_AllFail_ReturnsError(t *testing.T) {
	e := errors.New("fail")
	c, _ := NewChainClient(&staticErrorReader{err: e}, &staticErrorReader{err: e})
	_, err := c.ReadSecret(context.Background(), "secret/x")
	if err == nil {
		t.Fatal("expected error when all readers fail")
	}
}

func TestChainClient_NoReaders_ReturnsError(t *testing.T) {
	_, err := NewChainClient()
	if err == nil {
		t.Fatal("expected error constructing empty chain")
	}
}

func TestChainClient_Len(t *testing.T) {
	c, _ := NewChainClient(
		&staticDataReader{},
		&staticDataReader{},
		&staticDataReader{},
	)
	if c.Len() != 3 {
		t.Errorf("expected Len=3, got %d", c.Len())
	}
}

func TestChainClient_SingleReader_Success(t *testing.T) {
	r := &staticDataReader{data: map[string]interface{}{"only": true}}
	c, _ := NewChainClient(r)
	data, err := c.ReadSecret(context.Background(), "secret/only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["only"] != true {
		t.Errorf("unexpected data: %v", data)
	}
}
