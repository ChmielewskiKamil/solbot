package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"solbot/analyzer"
	"solbot/lsp"
	"solbot/lsp/analysis"
	"solbot/lsp/rpc"
	"solbot/parser"
	"solbot/reporter"
	"solbot/token"
)

func main() {
	mode := flag.String("mode", "analyzer", "Operation mode: lsp or analyzer")
	filePath := flag.String("file", "", "File path to analyze")
	flag.Parse()

	switch *mode {
	case "lsp":
		startLanguageServer()
	case "analyzer":
		if *filePath == "" {
			log.Fatalf("File path is required in analyzer mode.\nUse --file path/to/file.sol to analyze a file.")
		}
		startAnalyzer(*filePath)
	default:
		log.Fatalf("Unknown mode: `%s` Available modes: `lsp` or `analyzer`", *mode)
		os.Exit(1)
	}
}

func startLanguageServer() {
	logger := getLogger("log.txt")
	logger.Println("Logger started.")

	state := analysis.NewState()
	scanner := bufio.NewScanner(os.Stdin)
	writer := os.Stdout

	scanner.Split(rpc.Split)

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, content, err := rpc.DecodeMessage(msg)
		if err != nil {
			logger.Printf("Error decoding message: %s\n", err)
			continue
		}

		handleMessage(logger, writer, state, method, content)
	}
}

func startAnalyzer(filePath string) {
	println("Solbot starts")
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %s\n", err)
	}

	p := parser.Parser{}
	handle := token.NewFile(filePath, string(src))
	p.Init(handle)

	file := p.ParseFile()

	println("Solbot is analyzing your file...")
	findings := analyzer.AnalyzeFile(file)

	reporter.GenerateReport(findings, "solbot.md")
}

func handleMessage(logger *log.Logger, writer io.Writer, state analysis.State, method string, content []byte) {
	logger.Printf("Received message with method: %s\n", method)
	logger.Printf("Message content: %s\n", content)

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("initialize: %s\n", err)
			return
		}

		logger.Printf("Connected to: %s %s\n", request.Params.ClientInfo.Name, request.Params.ClientInfo.Version)

		msg := lsp.NewInitializeResponse(request.ID)
		writeResponse(writer, logger, msg)
	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("textDocument/didOpen: %s\n", err)
			return
		}

		logger.Printf("Opened: %s\n", request.Params.TextDocument.URI)

		// @TODO: Here we can start the static analysis

		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)
	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("textDocument/didChange: %s\n", err)
			return
		}

		logger.Printf("Changed: %s\n", request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}
	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("textDocument/hover: %s\n", err)
			return
		}

		response := state.Hover(request.ID, request.Params.TextDocument.URI, request.Params.Position)
		writeResponse(writer, logger, response)
	case "textDocument/definition":
		var request lsp.DefinitionRequest
		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("textDocument/definition: %s\n", err)
			return
		}

		response := state.Definition(request.ID, request.Params.TextDocument.URI, request.Params.Position)
		writeResponse(writer, logger, response)
	}
}

func writeResponse(writer io.Writer, logger *log.Logger, msg any) {
	response := rpc.EncodeMessage(msg)
	_, err := writer.Write([]byte(response))
	if err != nil {
		logger.Printf("Error writing response: %s\n", err)
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
