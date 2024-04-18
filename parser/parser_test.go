package parser

import (
	"solparsor/ast"
	"solparsor/token"
	"testing"
)

func Test_ParseElementaryTypes(t *testing.T) {
	src := `
    address owner = 0x12345;
    uint256 balance = 100;
    bool isOwner = true;
    `

	p := Parser{}
	p.init(src)

	file := p.ParseFile()

	if file == nil {
		t.Fatalf("ParseFile() returned nil")
	}

	if len(file.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations, got %d", len(file.Declarations))
	}

	tests := []struct {
		expectedType       token.TokenType
		expectedIdentifier string
	}{
		{token.ADDRESS, "owner"},
		{token.UINT_256, "balance"},
		{token.BOOL, "isOwner"},
	}

	for i, tt := range tests {
		decl := file.Declarations[i]
		if !testParseElementaryType(t, decl, tt.expectedType, tt.expectedIdentifier) {
			return
		}
	}
}

func testParseElementaryType(t *testing.T, decl ast.Declaration,
	expectedType token.TokenType, expectedIdentifier string) bool {
	if decl == nil {
		t.Fatalf("Expected Declaration, got nil")
	}

	vd, ok := decl.(*ast.VariableDeclaration)
	if !ok {
		t.Errorf("Expected VariableDeclaration, got %T", decl)
		return false
	}

	if vd.Name.Name != expectedIdentifier {
		t.Errorf("Expected identifier %s, got %s",
			expectedIdentifier, vd.Name.Name)
		return false
	}

	et, ok := vd.Type.(*ast.ElementaryType)
	if !ok {
		t.Errorf("Expected ElementaryType, got %T", vd.Type)
		return false
	}

	if et.Kind.Type != expectedType {
		t.Errorf("Expected token type %T, got %T", expectedType, et.Kind.Type)
		return false
	}

	return true
}
