package analysis

import (
	"fmt"
	"solbot/lsp"
)

type State struct {
	Documents map[string]string // file name -> file content
}

func NewState() State {
	return State{
		Documents: map[string]string{},
	}
}

func (s *State) OpenDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) UpdateDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) Hover(id int, uri string, position lsp.Position) lsp.HoverResponse {
	// @TODO: This should look up the type etc.

	_, ok := s.Documents[uri]
	if !ok {
		return lsp.NewHoverResponse(id, "")
	}

	content := fmt.Sprintf("Hover in file: %s, line: %d, character: %d", uri, position.Line, position.Character)

	return lsp.NewHoverResponse(id, content)
}

func (s *State) Definition(
	id int,
	uri string,
	position lsp.Position) lsp.DefinitionResponse {
	_, ok := s.Documents[uri]
	if !ok {
		return lsp.NewDefinitionResponse(id, nil)
	}

	locations := []lsp.Location{}

	loc1 := lsp.Location{
		URI: uri,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      position.Line + 3,
				Character: 0,
			},
			End: lsp.Position{
				Line:      position.Line + 3,
				Character: 0,
			},
		},
	}

	locations = append(locations, loc1)

	return lsp.NewDefinitionResponse(id, &locations)
}
