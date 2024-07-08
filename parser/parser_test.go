package parser

import (
	"solbot/ast"
	"solbot/token"
	"testing"
)

func Test_ParseElementaryTypes(t *testing.T) {
	src := `address owner = 0x12345;
    uint256 balance = 100;
    bool isOwner = true;
    bool constant IS_OWNER = true;           
    bool constant isOwner = false;            
    bool constant is_owner = false;           
    uint256 balance = 100;                    
    address constant router = 0x1337;         
    bool isOwner = true;                      
    uint16 constant ONE_hundred_IS_100 = 100; 
    uint256 constant DENOMINATOR = 1_000_000; 
    uint256 private constant Is_This_Snake_Case = 0;
    `

	p := Parser{}

	handle := token.NewFile("test.sol", src)
	p.Init(handle)

	file := p.ParseFile()
	checkParserErrors(t, &p)

	if file == nil {
		t.Fatalf("ParseFile() returned nil")
	}

	if len(file.Declarations) != 12 {
		t.Fatalf("Expected 12 declarations, got %d", len(file.Declarations))
	}

	tests := []struct {
		expectedType       token.TokenType
		expectedMutability ast.Mutability
		expectedVisibility ast.Visibility
		expectedIdentifier string
	}{
		{token.ADDRESS, 0, ast.Internal, "owner"},
		{token.UINT_256, 0, ast.Internal, "balance"},
		{token.BOOL, 0, ast.Internal, "isOwner"},
		{token.BOOL, 4, ast.Internal, "IS_OWNER"},
		{token.BOOL, 4, ast.Internal, "isOwner"},
		{token.BOOL, 4, ast.Internal, "is_owner"},
		{token.UINT_256, 0, ast.Internal, "balance"},
		{token.ADDRESS, 4, ast.Internal, "router"},
		{token.BOOL, 0, ast.Internal, "isOwner"},
		{token.UINT_16, 4, ast.Internal, "ONE_hundred_IS_100"},
		{token.UINT_256, 4, ast.Internal, "DENOMINATOR"},
		{token.UINT_256, 4, ast.Private, "Is_This_Snake_Case"},
	}

	for i, tt := range tests {
		decl := file.Declarations[i]
		if !testParseElementaryType(
			t,
			decl,
			tt.expectedType,
			tt.expectedMutability,
			tt.expectedVisibility,
			tt.expectedIdentifier) {
			return
		}
	}
}

func Test_ParseFunctionDeclaration(t *testing.T) {
	// @TODO: When overriding the param identifier can be empty?
	// e.g. function withdraw(uint256 assets, uint256) internal override ...
	src := `
    function getBalance(address owner) public view returns (uint256) {
        uint256 balance = 10;
        return balance;
    }
    `

	p := Parser{}
	handle := token.NewFile("test.sol", src)
	p.Init(handle)

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

	// if len(fd.Type.Params.List) != 1 {
	// 	t.Fatalf("Expected 1 parameter, got %d", len(fd.Type.Params.List))
	// }
	//
	// param := fd.Type.Params.List[0]
	// if param.Name.Name != "owner" {
	// 	t.Errorf("Expected parameter name owner, got %s", param.Name.Name)
	// }

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

	if fd.Body == nil {
		t.Fatalf("Expected BlockStatement, got nil")
	}

	// fb, ok := fd.Body.(*ast.BlockStatement)
	// if !ok {
	// 	t.Fatalf("Expected BlockStatement, got %T", fd.Body)
	// }

	// if len(fb.Statements) != 2 {
	// 	t.Fatalf("Expected 2 statements, got %d", len(fb.Statements))
	// }
}

func testParseElementaryType(t *testing.T, decl ast.Declaration,
	expectedType token.TokenType, expectedMutability ast.Mutability,
	expectedVisibility ast.Visibility, expectedIdentifier string) bool {
	if decl == nil {
		t.Fatalf("Expected Declaration, got nil")
	}

	vd, ok := decl.(*ast.StateVariableDeclaration)
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

	if vd.Visibility != expectedVisibility {
		t.Errorf("Expected ast visibility token type %v, got %v",
			expectedVisibility, vd.Visibility)
		return false
	}

	if vd.Mutability != expectedMutability {
		t.Errorf("Expected ast mutability token type %v, got %v", expectedMutability, vd.Mutability)
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
