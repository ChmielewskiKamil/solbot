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
	checkParserErrors(t, &p)

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

func Test_ParseFunctionDeclaration(t *testing.T) {
	src := `
    function getBalance(address owner) public view returns (uint256) {
        uint256 balance = 10;
        return balance;
    }
    `

	p := Parser{}
	p.init(src)

	file := p.ParseFile()
	checkParserErrors(t, &p)

	if file == nil {
		t.Fatalf("ParseFile() returned nil")
	}

	if len(file.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(file.Declarations))
	}

	decl := file.Declarations[0]
	fd, ok := decl.(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("Expected FunctionDeclaration, got %T", decl)
	}

	if fd.Name.Name != "getBalance" {
		t.Errorf("Expected function name getBalance, got %s", fd.Name.Name)
	}

	if fd.Type == nil {
		t.Fatalf("Expected FunctionType, got nil")
	}

	if fd.Type.Params == nil {
		t.Fatalf("Expected ParamList, got nil")
	}

	if len(fd.Type.Params.List) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(fd.Type.Params.List))
	}
	//
	param := fd.Type.Params.List[0]
	if param.Name.Name != "owner" {
		t.Errorf("Expected parameter name owner, got %s", param.Name.Name)
	}

	// @TODO: We skip the type for now since it is an expression.
	// if param.Type == nil {
	// 	t.Fatalf("Expected ElementaryType, got nil")
	// }

	// et, ok := param.Type.(*ast.ElementaryType)
	// if !ok {
	// 	t.Fatalf("Expected ElementaryType, got %T", param.Type)
	// }
	//
	// if et.Kind.Type != token.ADDRESS {
	// 	t.Errorf("Expected token type ADDRESS, got %T", et.Kind.Type)
	// }
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

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.errors
	if len(errors) == 0 {
		return
	}

	t.Errorf("Parser has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("Parser error: %s", err.Msg)
	}
	t.FailNow()
}
