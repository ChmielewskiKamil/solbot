package lsp

type InitializeRequest struct {
	Request
	Params InitializeParams `json:"params"`
}

type InitializeParams struct {
	// Since this is optional we can do a pointer.
	ClientInfo *ClientInfo `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync   int  `json:"textDocumentSync"` // Sync kind: 1 = full content, 2 = incremental
	HoverProvider      bool `json:"hoverProvider"`
	DefinitionProvider bool `json:"definitionProvider"` // Go to implementation of code that will be executed.
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewInitializeResponse(id int) InitializeResponse {
	return InitializeResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:   1, // Sync by sending the full content.
				HoverProvider:      true,
				DefinitionProvider: true,
			},
			ServerInfo: ServerInfo{
				Name:    "solbot_lsp",
				Version: "0.0.0-alpha",
			},
		},
	}
}
