package rpc

import (
	"testing"
)

type ExampleEncoding struct {
	Testing bool
}

func TestEncoding(t *testing.T) {
	msg := ExampleEncoding{Testing: true}
	encoded := EncodeMessage(msg)
	expected := "Content-Length: 16\r\n\r\n{\"Testing\":true}"
	if encoded != expected {
		t.Errorf("Expected %s, got %s", expected, encoded)
	}
}

func TestDecoding(t *testing.T) {
	msg := ExampleEncoding{Testing: true}
	encoded := EncodeMessage(msg)
	length, err := DecodeMessage([]byte(encoded))
	if err != nil {
		t.Fatalf("Error decoding message: %s", err)
	}

	expectedLen := 16

	if length != expectedLen {
		t.Fatalf("Expected length %d, got %d", expectedLen, length)
	}
}
