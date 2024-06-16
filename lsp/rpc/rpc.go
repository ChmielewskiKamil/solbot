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

type BaseMessage struct {
	Method string `json:"method"`
}

func DecodeMessage(msg []byte) (string, []byte, error) {
	header, content, found := bytes.Cut(msg, []byte("\r\n\r\n"))
	if !found {
		return "", nil, fmt.Errorf("Separator not found in message. Could not decode.")
	}

	contentLengthInBytes := header[len("Content-Length: "):]
	contentLength, err := strconv.Atoi(string(contentLengthInBytes))
	if err != nil {
		return "", nil, fmt.Errorf("Could not parse Content-Length: %s", err)
	}

	var message BaseMessage
	err = json.Unmarshal(content, &message)
	if err != nil {
		return "", nil, fmt.Errorf("Could not unmarshal message: %s", err)
	}

	return message.Method, content[:contentLength], nil
}

// Split is a function used for the bufio.Scanner to split the incoming data.
// For the LSP it will just split it based on the Content-Length header.
func Split(data []byte, _ bool) (advance int, token []byte, err error) {
	separator := []byte("\r\n\r\n")
	header, content, found := bytes.Cut(data, separator)
	// If not found yet, we don't want to error out, since the next chunk
	// of data might contain the separator.
	if !found {
		return 0, nil, nil
	}

	contentLengthInBytes := header[len("Content-Length: "):]
	contentLength, err := strconv.Atoi(string(contentLengthInBytes))
	// Here we return and error. If we can't get the actual number of bytes, something is messed up.
	if err != nil {
		return 0, nil, err
	}

	// We have to wait a moment for the rest of the data to arrive.
	if len(content) < contentLength {
		return 0, nil, nil
	}

	totalLength := len(header) + len(separator) + contentLength

	return totalLength, data[:totalLength], nil
}
