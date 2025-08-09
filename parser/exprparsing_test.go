package parser_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ChmielewskiKamil/solbot/ast"
	"github.com/ChmielewskiKamil/solbot/parser"
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

	test_Identifier(t, callExpr.Ident, "foo")

	if len(callExpr.Args) != 3 {
		t.Fatalf("Expected 3 arguments, got %d", len(callExpr.Args))
	}

	test_InfixExpression(t, callExpr.Args[0], "a", "+", "b")
	test_InfixExpression(t, callExpr.Args[1], big.NewInt(3), "*", big.NewInt(5))
	test_LiteralExpression(t, callExpr.Args[2], "bar")
}

func TestParseExpressions(t *testing.T) {
	// This helper function simplifies parsing expressions by wrapping them in a function
	// and extracting the statements, similar to the original test setup.
	parseExpressionStatements := func(t *testing.T, src string) []ast.Statement {
		// Wrap the expressions in a function for valid parsing
		fullSrc := "function wrapper() { " + src + " }"
		file := test_helper_parseSource(t, fullSrc, false)
		fnBody := test_helper_parseFnBody(t, file)
		return fnBody.Statements
	}

	testCases := []struct {
		name     string
		source   string
		validate func(t *testing.T, stmts []ast.Statement)
	}{
		{
			name:   "simple identifier",
			source: `foo;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 1 {
					t.Fatalf("Expected 1 statement, got %d", len(stmts))
				}
				exprStmt, ok := stmts[0].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("Expected ExpressionStatement, got %T", stmts[0])
				}
				test_LiteralExpression(t, exprStmt.Expression, "foo")
			},
		},
		{
			name:   "number literals",
			source: `1337; 0x12345;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 2 {
					t.Fatalf("Expected 2 statements, got %d", len(stmts))
				}
				expectedValues := []*big.Int{big.NewInt(1337), big.NewInt(0x12345)}
				for i, stmt := range stmts {
					exprStmt, ok := stmt.(*ast.ExpressionStatement)
					if !ok {
						t.Fatalf("Statement %d: Expected ExpressionStatement, got %T", i, stmt)
					}
					test_LiteralExpression(t, exprStmt.Expression, expectedValues[i])
				}
			},
		},
		{
			name:   "boolean literals",
			source: `true; false;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 2 {
					t.Fatalf("Expected 2 statements, got %d", len(stmts))
				}
				expectedValues := []bool{true, false}
				for i, stmt := range stmts {
					exprStmt, ok := stmt.(*ast.ExpressionStatement)
					if !ok {
						t.Fatalf("Statement %d: Expected ExpressionStatement, got %T", i, stmt)
					}
					test_LiteralExpression(t, exprStmt.Expression, expectedValues[i])
				}
			},
		},
		{
			name:   "prefix expressions",
			source: `-1337; !a; delete foo;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 3 {
					t.Fatalf("Expected 3 statements, got %d", len(stmts))
				}
				tests := []struct {
					operator    string
					expectedVal interface{}
				}{
					{"-", big.NewInt(1337)},
					{"!", "a"},
					{"delete", "foo"},
				}
				for i, tt := range tests {
					exprStmt := stmts[i].(*ast.ExpressionStatement)
					pExpr, ok := exprStmt.Expression.(*ast.PrefixExpression)
					if !ok {
						t.Fatalf("Expected PrefixExpression, got %T", exprStmt.Expression)
					}
					if pExpr.Operator.Literal != tt.operator {
						t.Fatalf("Expected operator %s, got %s", tt.operator, pExpr.Operator.Literal)
					}
					test_LiteralExpression(t, pExpr.Right, tt.expectedVal)
				}
			},
		},
		{
			name:   "infix expressions",
			source: `a + b; true == false;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 2 {
					t.Fatalf("Expected 2 statements, got %d", len(stmts))
				}
				test_InfixExpression(t, stmts[0].(*ast.ExpressionStatement).Expression, "a", "+", "b")
				test_InfixExpression(t, stmts[1].(*ast.ExpressionStatement).Expression, true, "==", false)
			},
		},
		{
			name:   "operator precedence",
			source: `a + b * c; -a * b; 3 * (8 + 2) * 2; foo(a * b) + bar(c / d, e);`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 4 {
					t.Fatalf("Expected 4 statements, got %d", len(stmts))
				}
				expectedStrings := []string{
					"(a + (b * c))",
					"((-a) * b)",
					"((3 * (8 + 2)) * 2)",
					"(foo((a * b)) + bar((c / d), e))",
				}
				for i, expected := range expectedStrings {
					actual := stmts[i].(*ast.ExpressionStatement).Expression.String()
					if actual != expected {
						t.Errorf("Statement %d: expected '%s', got '%s'", i, expected, actual)
					}
				}
			},
		},
		{
			name:   "call expression with complex arguments",
			source: `foo(a + b, 3 * 5, bar);`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 1 {
					t.Fatalf("Expected 1 statement, got %d", len(stmts))
				}
				exprStmt := stmts[0].(*ast.ExpressionStatement)
				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("Expected CallExpression, got %T", exprStmt.Expression)
				}
				test_Identifier(t, callExpr.Ident, "foo")
				if len(callExpr.Args) != 3 {
					t.Fatalf("Expected 3 arguments, got %d", len(callExpr.Args))
				}
				test_InfixExpression(t, callExpr.Args[0], "a", "+", "b")
				test_InfixExpression(t, callExpr.Args[1], big.NewInt(3), "*", big.NewInt(5))
				test_LiteralExpression(t, callExpr.Args[2], "bar")
			},
		},
		{
			name:   "member access and call chain",
			source: `a.b.c(d, e.f);`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 1 {
					t.Fatalf("Expected 1 statement, got %d", len(stmts))
				}
				// Expected AST: Call( MemberAccess( MemberAccess(a, b), c ), [d, MemberAccess(e, f)] )
				exprStmt := stmts[0].(*ast.ExpressionStatement)
				callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("Expected top-level expression to be a CallExpression, got %T", exprStmt.Expression)
				}

				// Validate the function being called: a.b.c
				memberAccess1, ok := callExpr.Ident.(*ast.MemberAccessExpression)
				if !ok {
					t.Fatalf("Expected call identifier to be a MemberAccessExpression, got %T", callExpr.Ident)
				}
				test_Identifier(t, memberAccess1.Member, "c")

				memberAccess2, ok := memberAccess1.Expression.(*ast.MemberAccessExpression)
				if !ok {
					t.Fatalf("Expected nested expression to be a MemberAccessExpression, got %T", memberAccess1.Expression)
				}
				test_Identifier(t, memberAccess2.Member, "b")
				test_Identifier(t, memberAccess2.Expression, "a")

				// Validate the arguments: [d, e.f]
				if len(callExpr.Args) != 2 {
					t.Fatalf("Expected 2 arguments, got %d", len(callExpr.Args))
				}
				test_Identifier(t, callExpr.Args[0], "d")
				arg2, ok := callExpr.Args[1].(*ast.MemberAccessExpression)
				if !ok {
					t.Fatalf("Expected second argument to be a MemberAccessExpression, got %T", callExpr.Args[1])
				}
				test_Identifier(t, arg2.Member, "f")
				test_Identifier(t, arg2.Expression, "e")
			},
		},
		{
			name:   "postfix expressions",
			source: `i++; j--; k++ * 2;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 3 {
					t.Fatalf("Expected 3 statements, got %d", len(stmts))
				}

				// Test case 1: i++
				stmt1 := stmts[0].(*ast.ExpressionStatement)
				postfix1, ok := stmt1.Expression.(*ast.PostfixExpression)
				if !ok {
					t.Fatalf("Stmt 1: Expected PostfixExpression, got %T", stmt1.Expression)
				}
				if postfix1.Operator.Literal != "++" {
					t.Errorf("Stmt 1: Expected operator ++, got %s", postfix1.Operator.Literal)
				}
				test_Identifier(t, postfix1.Left, "i")

				// Test case 2: j--
				stmt2 := stmts[1].(*ast.ExpressionStatement)
				postfix2, ok := stmt2.Expression.(*ast.PostfixExpression)
				if !ok {
					t.Fatalf("Stmt 2: Expected PostfixExpression, got %T", stmt2.Expression)
				}
				if postfix2.Operator.Literal != "--" {
					t.Errorf("Stmt 2: Expected operator --, got %s", postfix2.Operator.Literal)
				}
				test_Identifier(t, postfix2.Left, "j")

				// Test case 3: k++ * 2 (tests precedence)
				// Expected AST: Infix( Postfix(k, ++), *, 2 )
				stmt3 := stmts[2].(*ast.ExpressionStatement)
				infix, ok := stmt3.Expression.(*ast.InfixExpression)
				if !ok {
					t.Fatalf("Stmt 3: Expected InfixExpression, got %T", stmt3.Expression)
				}
				if infix.Operator.Literal != "*" {
					t.Errorf("Stmt 3: Expected operator *, got %s", infix.Operator.Literal)
				}

				// Check left side of the multiplication
				postfix3, ok := infix.Left.(*ast.PostfixExpression)
				if !ok {
					t.Fatalf("Stmt 3: Expected left side to be PostfixExpression, got %T", infix.Left)
				}
				if postfix3.Operator.Literal != "++" {
					t.Errorf("Stmt 3: Expected postfix operator ++, got %s", postfix3.Operator.Literal)
				}
				test_Identifier(t, postfix3.Left, "k")

				// Check right side of the multiplication
				test_NumberLiteral(t, infix.Right, big.NewInt(2))
			},
		},
		{
			name:   "postfix binds tighter than prefix",
			source: `!i++;`,
			validate: func(t *testing.T, stmts []ast.Statement) {
				if len(stmts) != 1 {
					t.Fatalf("Expected 1 statement, got %d", len(stmts))
				}

				// The expected structure is: Prefix( !, Postfix(i, ++) )
				// The string representation should be "(!(i++))"
				exprStmt := stmts[0].(*ast.ExpressionStatement)

				// 1. Check the root is a PrefixExpression (!)
				prefixExpr, ok := exprStmt.Expression.(*ast.PrefixExpression)
				if !ok {
					t.Fatalf("Expected root to be PrefixExpression, got %T", exprStmt.Expression)
				}
				if prefixExpr.Operator.Literal != "!" {
					t.Errorf("Expected prefix operator to be '!', got '%s'", prefixExpr.Operator.Literal)
				}

				// 2. Check the right side of the '!' is a PostfixExpression (i++)
				postfixExpr, ok := prefixExpr.Right.(*ast.PostfixExpression)
				if !ok {
					t.Fatalf("Expected right side of prefix to be PostfixExpression, got %T", prefixExpr.Right)
				}
				if postfixExpr.Operator.Literal != "++" {
					t.Errorf("Expected postfix operator to be '++', got '%s'", postfixExpr.Operator.Literal)
				}

				// 3. Check the left side of the '++' is the identifier 'i'
				test_Identifier(t, postfixExpr.Left, "i")

				// 4. Finally, check the string output for a definitive tree structure check
				expectedStr := "(!(i++))"
				if exprStmt.Expression.String() != expectedStr {
					t.Errorf("Expected string representation '%s', got '%s'", expectedStr, exprStmt.Expression.String())
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use the helper to parse the source string
			statements := parseExpressionStatements(t, tc.source)
			// Run the custom validation logic for this test case
			tc.validate(t, statements)
		})
	}
}

/*~*~*~*~*~*~*~*~*~*~*~*~* Helper Functions ~*~*~*~*~*~*~*~*~*~*~*~*~*/

func test_helper_parseSource(t *testing.T, src string, tracing bool) *ast.File {
	var file *ast.File
	var err error

	if tracing {
		// Call ParseFile with the tracing option.
		file, err = parser.ParseFile("test_file.sol", strings.NewReader(src), parser.WithTracing())
	} else {
		// Call it without the option.
		file, err = parser.ParseFile("test_file.sol", strings.NewReader(src))
	}

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

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
