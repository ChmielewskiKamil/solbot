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

func Test_ParseBooleanLiteralExpression(t *testing.T) {
	src := `function test() public {
        true;
        false;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(fnBody.Statements))
	}

	tests := []struct {
		expectedVal bool
	}{
		{true},
		{false},
	}

	for i, tt := range tests {
		expr := fnBody.Statements[i]
		exprStmt, ok := expr.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", expr)
		}

		boolLit, ok := exprStmt.Expression.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("Expected BooleanLiteral, got %T", exprStmt.Expression)
		}

		if boolLit.Value != tt.expectedVal {
			t.Fatalf("Expected %t, got %t", tt.expectedVal, boolLit.Value)
		}
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
        !true;
        !false;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 8 {
		t.Fatalf("Expected 8 statements, got %d", len(fnBody.Statements))
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
		{"!", true},
		{"!", false},
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
	}
}

func Test_ParseInfixExpressions(t *testing.T) {
	src := `function test() public {
        2 + 2;
        2 - 2;
        2 * 2;
        2 / 2;
        a + b;
        true == true;
        false != true;
        false == false;
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 8 {
		t.Fatalf("Expected 8 statements, got %d", len(fnBody.Statements))
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
		{true, "==", true},
		{false, "!=", true},
		{false, "==", false},
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
        3 > 8 == false;
        3 < 8 == true;
        3 * (8 + 2) * 2;
        10 / (1 + 1);
        a + foo(b + c);
        foo(a * b, c / d + e)
        foo(a * b) + bar(c / d, e);
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 23 {
		t.Fatalf("Expected 20 statements, got %d", len(fnBody.Statements))
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
		{"((3 > 8) == false)"},
		{"((3 < 8) == true)"},
		{"((3 * (8 + 2)) * 2)"},
		{"(10 / (1 + 1))"},
		{"(a + foo((b + c)))"},
		{"foo((a * b), ((c / d) + e))"},
		{"(foo((a * b)) + bar((c / d), e))"},
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

func Test_ParseCallExpression(t *testing.T) {
	src := `function test() public {
        foo(a + b, 3 * 5, bar);
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(fnBody.Statements))
	}

	exprStmt, ok := fnBody.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", fnBody.Statements[0])
	}

	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %T", exprStmt.Expression)
	}

	test_Identifier(t, callExpr.Function, "foo")

	if len(callExpr.Args) != 3 {
		t.Fatalf("Expected 3 arguments, got %d", len(callExpr.Args))
	}

	test_InfixExpression(t, callExpr.Args[0], "a", "+", "b")
	test_InfixExpression(t, callExpr.Args[1], big.NewInt(3), "*", big.NewInt(5))
	test_LiteralExpression(t, callExpr.Args[2], "bar")
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

	if ident.Value != value {
		t.Fatalf("Expected %s, got %s", value, ident.Value)
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

func test_BooleanLiteral(t *testing.T, exp ast.Expression, value bool) {
	bl, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("Expected BooleanLiteral, got %T", exp)
	}

	if bl.Value != value {
		t.Fatalf("Expected %t, got %t", value, bl.Value)
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
	case bool:
		test_BooleanLiteral(t, exp, v)
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
