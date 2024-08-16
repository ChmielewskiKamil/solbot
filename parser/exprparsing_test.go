package parser

import (
	"math/big"
	"solbot/ast"
	"solbot/token"
	"testing"
)

func Test_ParseIdentifierExpression(t *testing.T) {
	src := `function test() public {
        foo;
    }`

	p := Parser{}
	handle := token.NewFile("test.sol", src)
	p.Init(handle)
	// p.ToggleTracing()

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

	if fd.Body == nil {
		t.Fatalf("FunctionDeclaration body is nil")
	}

	if len(fd.Body.Statements) != 1 {
		t.Fatalf("Expected 1 statements, got %d", len(fd.Body.Statements))
	}

	stmt := fd.Body.Statements[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", stmt)
	}

	ident, ok := exprStmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier, got %T", exprStmt.Expression)
	}

	if ident.Name != "foo" {
		t.Fatalf("Expected foo, got %s", ident.Name)
	}
}

func Test_ParseNumberLiteralExpression(t *testing.T) {
	src := `function test() public {
        1337;
        0x12345;
        0x000000;
        115792089237316195423570985008687907853269984665640564039457584007913129639935;
    }`

	p := Parser{}
	handle := token.NewFile("test.sol", src)
	p.Init(handle)
	p.ToggleTracing()

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

	if fd.Body == nil {
		t.Fatalf("FunctionDeclaration body is nil")
	}

	if len(fd.Body.Statements) != 4 {
		t.Fatalf("Expected 4 statements, got %d", len(fd.Body.Statements))
	}

	uint256max, _ := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 0)

	tests := []struct {
		expectedVal     *big.Int
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{big.NewInt(1337), token.DECIMAL_NUMBER, "1337"},
		{big.NewInt(0x12345), token.HEX_NUMBER, "0x12345"},
		{big.NewInt(0x000000), token.HEX_NUMBER, "0x000000"},
		{uint256max, token.DECIMAL_NUMBER, "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
	}

	for i, tt := range tests {
		expr := fd.Body.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		testNumberLiteral(t, exprStmt.Expression, tt.expectedVal,
			tt.expectedType, tt.expectedLiteral)
	}
}

func testNumberLiteral(
	t *testing.T,
	expr ast.Expression,
	expectedVal *big.Int,
	expectedType token.TokenType,
	expectedLiteral string) {
	intLit, ok := expr.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("Expected IntegerLiteral, got %T", expr)
	}

	if intLit.Kind.Literal != expectedLiteral {
		t.Fatalf("Expected %s, got %s", expectedLiteral, intLit.Kind.Literal)
	}

	if intLit.Value.Cmp(expectedVal) != 0 {
		t.Fatalf("Expected %d, got %s", expectedVal, intLit.Value.String())
	}

	if intLit.Kind.Type != expectedType {
		t.Fatalf("Expected %s, got %s", expectedType, intLit.Kind.Type)
	}
}
