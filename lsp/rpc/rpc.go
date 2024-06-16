package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func EncodeMessage(msg any) string {
	content, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	// This conforms to the LSP specification.
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(content), content)
}

func DecodeMessage(msg []byte) (int, error) {
	header, content, found := bytes.Cut(msg, []byte("\r\n\r\n"))
	if !found {
		return 0, fmt.Errorf("Separator not found in message. Could not decode.")
	}

	contentLengthInBytes := header[len("Content-Length: "):]
	contentLength, err := strconv.Atoi(string(contentLengthInBytes))
	if err != nil {
		return 0, fmt.Errorf("Could not parse Content-Length: %s", err)
	}

	// @TODO: Handle this later.
	_ = content

	return contentLength, nil
}
