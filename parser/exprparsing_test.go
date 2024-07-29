package parser

import (
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
