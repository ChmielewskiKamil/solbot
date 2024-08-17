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

func Test_ParsePrefixExpression(t *testing.T) {
	src := `function test() public {
        -1337;
        ++5;
        --123;
        ~0x12345;
        !true;
        delete foo;
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

	if len(fd.Body.Statements) != 6 {
		t.Fatalf("Expected 6 statements, got %d", len(fd.Body.Statements))
	}

	tests := []struct {
		operator        string
		expectedVal     *big.Int
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{"-", big.NewInt(1337), token.DECIMAL_NUMBER, "1337"},
		{"++", big.NewInt(5), token.DECIMAL_NUMBER, "5"},
		{"--", big.NewInt(123), token.DECIMAL_NUMBER, "123"},
		{"~", big.NewInt(0x12345), token.HEX_NUMBER, "0x12345"},
		{"!", nil, token.TRUE_LITERAL, "true"},
		{"delete", nil, token.IDENTIFIER, "foo"},
	}

	for i, tt := range tests {
		expr := fd.Body.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		pExpr, ok := exprStmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("Expected PrefixExpression, got %T", exprStmt.Expression)
		}

		if pExpr.Operator.Literal != tt.operator {
			t.Fatalf("Expected operator %s, got %s", tt.operator, pExpr.Operator)
		}

		if i < 4 {
			testNumberLiteral(t, pExpr.Right, tt.expectedVal, tt.expectedType, tt.expectedLiteral)
		}

		// @TODO: Implement tests for non number literals.
	}
}

func Test_ParseInfixExpressions(t *testing.T) {
	src := `function test() public {
        2 + 2;
        2 - 2;
        2 * 2;
        2 / 2;
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

	if len(fd.Body.Statements) != 4 {
		t.Fatalf("Expected 4 statements, got %d", len(fd.Body.Statements))
	}

	infixTests := []struct {
		leftVal  *big.Int
		operator string
		rightVal *big.Int
	}{
		{big.NewInt(2), "+", big.NewInt(2)},
		{big.NewInt(2), "-", big.NewInt(2)},
		{big.NewInt(2), "*", big.NewInt(2)},
		{big.NewInt(2), "/", big.NewInt(2)},
	}

	for i, tt := range infixTests {
		expr := fd.Body.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		infixExpr, ok := exprStmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("Expected InfixExpression, got %T", exprStmt.Expression)
		}

		testNumberLiteral(t, infixExpr.Left, tt.leftVal, token.DECIMAL_NUMBER, "2")

		if infixExpr.Operator.Literal != tt.operator {
			t.Fatalf("Expected operator %s, got %s", tt.operator, infixExpr.Operator)
		}

		testNumberLiteral(t, infixExpr.Right, tt.rightVal, token.DECIMAL_NUMBER, "2")
	}
}

func Test_ParseOperatorPrecedence(t *testing.T) {
	src := `function test() public {
        a + b;
        a - b;
        a * b;
        a / b;
        -a * b;
        a + b + c;
        a + b * c;
        -a - b;
        -a - -b;
        a + b * c + d / e - f;
        -a + -b; -a * -b; -a ** -b;
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

	if len(fd.Body.Statements) != 13 {
		t.Fatalf("Expected 13 statements, got %d", len(fd.Body.Statements))
	}

	tests := []struct {
		expected string
	}{
		{"(a + b)"},
		{"(a - b)"},
		{"(a * b)"},
		{"(a / b)"},
		{"((-a) * b)"},
		{"((a + b) + c)"},
		{"(a + (b * c))"},
		{"((-a) - b)"},
		{"((-a) - (-b))"},
		{"(((a + (b * c)) + (d / e)) - f)"},
		{"((-a) + (-b))"},
		{"((-a) * (-b))"},
		{"((-a) ** (-b))"},
	}

	for i, tt := range tests {
		expr := fd.Body.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		if exprStmt.String() != tt.expected {
			t.Fatalf("Expected %s, got %s", tt.expected, exprStmt.String())
		}
	}
}
