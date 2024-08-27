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

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 1 {
		t.Fatalf("Expected 1 statements, got %d", len(fnBody.Statements))
	}

	stmt := fnBody.Statements[0]
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", stmt)
	}

	test_LiteralExpression(t, exprStmt.Expression, "foo")
}

func Test_ParseNumberLiteralExpression(t *testing.T) {
	src := `function test() public {
        1337;
        0x12345;
        0x000000;
        115792089237316195423570985008687907853269984665640564039457584007913129639935;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 4 {
		t.Fatalf("Expected 4 statements, got %d", len(fnBody.Statements))
	}

	uint256max, _ := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 0)

	tests := []struct {
		expectedVal *big.Int
	}{
		{big.NewInt(1337)},
		{big.NewInt(0x12345)},
		{big.NewInt(0x000000)},
		{uint256max},
	}

	for i, tt := range tests {
		expr := fnBody.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		test_LiteralExpression(t, exprStmt.Expression, tt.expectedVal)
	}
}

func Test_ParsePrefixExpression(t *testing.T) {
	src := `function test() public {
        -1337;
        ++5;
        --123;
        ~0x12345;
        !a;
        delete foo;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 6 {
		t.Fatalf("Expected 6 statements, got %d", len(fnBody.Statements))
	}

	tests := []struct {
		operator    string
		expectedVal interface{}
	}{
		{"-", big.NewInt(1337)},
		{"++", big.NewInt(5)},
		{"--", big.NewInt(123)},
		{"~", big.NewInt(0x12345)},
		{"!", "a"},
		{"delete", "foo"},
		// {"!", nil, token.TRUE_LITERAL, "true"},
		// {"delete", nil, token.IDENTIFIER, "foo"},
	}

	for i, tt := range tests {
		expr := fnBody.Statements[i]
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

		test_LiteralExpression(t, pExpr.Right, tt.expectedVal)

		// @TODO: Implement tests for tokens different than
		// numbers and identifiers e.g. TRUE_LITERAL (true).
	}
}

func Test_ParseInfixExpressions(t *testing.T) {
	src := `function test() public {
        2 + 2;
        2 - 2;
        2 * 2;
        2 / 2;
        a + b;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 5 {
		t.Fatalf("Expected 5 statements, got %d", len(fnBody.Statements))
	}

	infixTests := []struct {
		leftVal  interface{}
		operator string
		rightVal interface{}
	}{
		{big.NewInt(2), "+", big.NewInt(2)},
		{big.NewInt(2), "-", big.NewInt(2)},
		{big.NewInt(2), "*", big.NewInt(2)},
		{big.NewInt(2), "/", big.NewInt(2)},
		{"a", "+", "b"},
	}

	for i, tt := range infixTests {
		expr := fnBody.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		test_InfixExpression(t, exprStmt.Expression, tt.leftVal, tt.operator, tt.rightVal)
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
        ++a;
        ++a + ++b;
        1 + 2;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 16 {
		t.Fatalf("Expected 16 statements, got %d", len(fnBody.Statements))
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
		{"(++a)"},
		{"((++a) + (++b))"},
		{"(1 + 2)"},
	}

	for i, tt := range tests {
		expr := fnBody.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		if exprStmt.String() != tt.expected {
			t.Fatalf("Expected %s, got %s", tt.expected, exprStmt.String())
		}
	}
}

/*~*~*~*~*~*~*~*~*~*~*~*~* Helper Functions ~*~*~*~*~*~*~*~*~*~*~*~*~*/

func test_helper_parseSource(t *testing.T, src string, tracing bool) *ast.File {
	p := Parser{}
	handle := token.NewFile("test.sol", src)
	p.Init(handle)

	if tracing {
		p.ToggleTracing()
	}

	file := p.ParseFile()
	checkParserErrors(t, &p)

	if file == nil {
		t.Fatalf("ParseFile() returned nil")
	}

	return file
}

func test_helper_parseFnBody(t *testing.T, file *ast.File) *ast.BlockStatement {
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

	return fd.Body
}

func test_Identifier(t *testing.T, exp ast.Expression, value string) {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier, got %T", exp)
	}

	if ident.Name != value {
		t.Fatalf("Expected %s, got %s", value, ident.Name)
	}
}

func test_NumberLiteral(
	t *testing.T,
	expr ast.Expression,
	expectedVal *big.Int) {
	intLit, ok := expr.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("Expected IntegerLiteral, got %T", expr)
	}

	if intLit.Value.Cmp(expectedVal) != 0 {
		t.Fatalf("Expected %d, got %s", expectedVal, intLit.Value.String())
	}
}

func test_LiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) {
	switch v := expected.(type) {
	case string:
		test_Identifier(t, exp, v)
		return
	case *big.Int:
		test_NumberLiteral(t, exp, v)
		return
	}
	t.Fatalf("Type %T not handled", expected)
}

func test_InfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) {
	infix, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("Expected InfixExpression, got %T", exp)
	}

	test_LiteralExpression(t, infix.Left, left)

	if infix.Operator.Literal != operator {
		t.Fatalf("Expected operator %s, got %s", operator, infix.Operator.Literal)
	}

	test_LiteralExpression(t, infix.Right, right)
}
