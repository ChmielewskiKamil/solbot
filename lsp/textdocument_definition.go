package lsp

type DefinitionRequest struct {
	Request
	Params DefinitionParams `json:"params"`
}

type DefinitionParams struct {
	TextDocumentPositionParams
}

type DefinitionResponse struct {
	Response
	// It works perfectly fine with a single location. For multiple locations
	// it lets the user to choose.
	Result *[]Location `json:"result,omitempty"`
}

func NewDefinitionResponse(id int, locations *[]Location) DefinitionResponse {
	return DefinitionResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: locations,
	}
}
