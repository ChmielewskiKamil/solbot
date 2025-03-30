package lsp

type HoverRequest struct {
	Request
	Params HoverParams `json:"params"`
}

type HoverParams struct {
	TextDocumentPositionParams
}

type HoverResponse struct {
	Response
	Result HoverResult `json:"result"`
}

type HoverResult struct {
	// TODO: This should be changed to MarkupContent
	Contents string `json:"contents"`
}

func NewHoverResponse(id int, contents string) HoverResponse {
	return HoverResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: HoverResult{
			Contents: contents,
		},
	}
}
