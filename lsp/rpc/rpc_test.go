package rpc

import (
	"testing"
)

type ExampleEncoding struct {
	Method string
}

var methodStr = "Hello world!"

func TestEncoding(t *testing.T) {
	msg := ExampleEncoding{Method: methodStr}
	encoded := EncodeMessage(msg)
	expected := "Content-Length: 25\r\n\r\n{\"Method\":\"Hello world!\"}"
	if encoded != expected {
		t.Errorf("Expected %s, got %s", expected, encoded)
	}
}

func TestDecoding(t *testing.T) {
	msg := ExampleEncoding{Method: methodStr}
	encoded := EncodeMessage(msg)
	decodedMethod, content, err := DecodeMessage([]byte(encoded))
	if err != nil {
		t.Fatalf("Error decoding message: %s", err)
	}

	length := len(content)
	expectedLen := 25

	if length != expectedLen {
		t.Fatalf("Expected length %d, got %d", expectedLen, length)
	}

	if decodedMethod != "Hello world!" {
		t.Fatalf("Expected method %s, got %s", methodStr, decodedMethod)
	}
}
