package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"solbot/lsp"
	"solbot/lsp/analysis"
	"solbot/lsp/rpc"
)

func main() {
	logger := getLogger("log.txt")
	logger.Println("Logger started.")

	state := analysis.NewState()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.Split)
	for scanner.Scan() {
		msg := scanner.Bytes()
		method, content, err := rpc.DecodeMessage(msg)
		if err != nil {
			logger.Printf("Error decoding message: %s\n", err)
			continue
		}

		handleMessage(logger, state, method, content)
	}
}

func handleMessage(logger *log.Logger, state analysis.State, method string, content []byte) {
	logger.Printf("Received message with method: %s\n", method)
	logger.Printf("Message content: %s\n", content)

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("initialize: %s\n", err)
		}

		logger.Printf("Connected to: %s %s\n", request.Params.ClientInfo.Name, request.Params.ClientInfo.Version)

		msg := lsp.NewInitializeResponse(request.ID)
		response := rpc.EncodeMessage(msg)
		_, err := os.Stdout.Write([]byte(response))
		if err != nil {
			logger.Printf("Error writing response: %s\n", err)
		}

		logger.Println("Sent initialize response.")
	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("textDocument/didOpen: %s\n", err)
		}

		logger.Printf("Opened: %s\n", request.Params.TextDocument.URI)
		// @TODO: Here we can start the static analysis
		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)
	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("textDocument/didChange: %s\n", err)
		}

		logger.Printf("Changed: %s\n", request.Params.TextDocument.URI)

		// @TODO: Here we can start the static analysis

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}
	}
}

func getLogger(filename string) *log.Logger {
	// Bitwise OR is used to combine the flags e.g. 001 | 010 | 100 is 111
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	return log.New(logfile, "[solbot_lsp] ", log.Ldate|log.Ltime|log.Lshortfile)
}
