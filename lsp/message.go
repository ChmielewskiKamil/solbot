package lsp

type Request struct {
	RPC    string `json:"jsonrpc"` // Useless, but we have to send it either way.
	ID     int    `json:"id"`
	Method string `json:"method"`
}

type Response struct {
	RPC string `json:"jsonrpc"` // Useless, but we have to send it either way.
	ID  *int   `json:"id,omitempty"`
}

type Notification struct {
	RPC    string `json:"jsonrpc"` // Useless, but we have to send it either way.
	Method string `json:"method"`
}
