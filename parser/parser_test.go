package parser

import (
	"solbot/ast"
	"solbot/token"
	"testing"
)

func Test_ParseStateVariableDeclaration(t *testing.T) {
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
    uint256 transient blob = 0;
    `

	file := test_helper_parseSource(t, src, false)

	if len(file.Declarations) != 13 {
		t.Fatalf("Expected 13 declarations, got %d", len(file.Declarations))
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
		{token.UINT_256, 6, ast.Internal, "blob"},
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
    function getBalance(address owner, uint256 amount) public view returns (uint256) {
        uint256 balance = 10;
        return balance;
    }
    `

	file := test_helper_parseSource(t, src, false)

	if len(file.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(file.Declarations))
	}

	decl := file.Declarations[0]
	fd, ok := decl.(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("Expected FunctionDeclaration, got %T", decl)
	}

	if fd.Name.Value != "getBalance" {
		t.Errorf("Expected function name getBalance, got %s", fd.Name.Value)
	}

	if fd.Params == nil {
		t.Fatalf("Expected ParamList, got nil")
	}

	if len(fd.Params.List) != 2 {
		t.Fatalf("Expected 2 parameter, got %d", len(fd.Params.List))
	}

	tests := []struct {
		expectedType       interface{}
		expectedIdentifier string
	}{
		{token, "owner"},
	}

	for i, tt := range tests {
		param := fd.Params.List[i]
		if param.Name.Value != tt.expectedIdentifier {
			t.Errorf("Expected parameter name %s, got %s", tt.expectedIdentifier, param.Name.Value)
		}

		if param.Type == nil {
			t.Fatalf("Expected ElementaryType, got nil")
		}

		et, ok := param.Type.(*ast.ElementaryType)
		if !ok {
			t.Fatalf("Expected ElementaryType, got %T", param.Type)
		}

		if et != tt.expectedType {
			t.Errorf("Expected token type %T, got %T", tt.expectedType, et)
		}

	}

	if fd.Body == nil {
		t.Fatalf("Expected BlockStatement, got nil")
	}

	if len(fd.Body.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(fd.Body.Statements))
	}

	stmt := fd.Body.Statements[0]
	vdStmt, ok := stmt.(*ast.VariableDeclarationStatement)
	if !ok {
		t.Fatalf("Expected VariableDeclarationStatement, got %T", stmt)
	}

	typ, ok := vdStmt.Type.(*ast.ElementaryType)
	if !ok {
		t.Fatalf("Expected ElementaryType, got %T", vdStmt.Type)
	}

	if typ.Kind.Type != token.UINT_256 {
		t.Errorf("Expected token type UINT_256, got %s", typ.Kind.Type.String())
	}

	if vdStmt.Name == nil {
		t.Fatalf("Expected Identifier, got nil")
	}

	if vdStmt.Name.Value != "balance" {
		t.Errorf("Expected balance, got %s", vdStmt.Name.Value)
	}

	if vdStmt.DataLocation != ast.NO_DATA_LOCATION {
		t.Fatalf("Expected NO_DATA_LOCATION, got %T", vdStmt.DataLocation)
	}

	stmt = fd.Body.Statements[1]
	_, ok = stmt.(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", stmt)
	}
}

// Since the return statement is a "statement", and there are no free-floating
// statements in Solidity, we have to wrap it in some kind of declaration e.g.
// a function declaration. Since the ast.File got a list of declarations, we
// can have 1 function decl and inside it test multiple test cases for return
// statements.
func Test_ParseReturnStatement(t *testing.T) {
	src := `function test() public {
    return 10;
    return 0x12345;
    return true;
    return staked;
    return address(0);
    return uint256(a + b);
    }
    `

	numReturns := 6

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != numReturns {
		t.Fatalf("Expected %d statements, got %d", numReturns, len(fnBody.Statements))
	}

	for _, stmt := range fnBody.Statements {
		retStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("Expected ReturnStatement, got %T", stmt)
		}

		// @TODO: Test the expression inside the return statement.
		_ = retStmt
	}
}

func Test_ParseBlocks(t *testing.T) {
	src := `function test() public {
    return 0x12345;
    return address(0);
    unchecked {
            return 0x12345;
        }
    return staked;
    }
    `

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 4 {
		t.Fatalf("Expected 4 statements, got %d", len(fnBody.Statements))
	}

	uncheckedBlock := fnBody.Statements[2]
	_, ok := uncheckedBlock.(*ast.UncheckedBlockStatement)
	if !ok {
		t.Fatalf("Expected UncheckedStatement, got %T", uncheckedBlock)
	}
}

func Test_ParseIfStatement(t *testing.T) {
	src := `function test() public {
        if (a > b) { return a; }
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(fnBody.Statements))
	}

	stmt := fnBody.Statements[0]
	ifStmt, ok := stmt.(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", stmt)
	}

	if ifStmt.Condition == nil {
		t.Fatalf("Condition expected to be <<Expression>>, got nil")
	}

	test_InfixExpression(t, ifStmt.Condition, "a", ">", "b")

	if ifStmt.Consequence == nil {
		t.Fatalf("Consequence expected to be <<BlockStatement>>, got nil")
	}

	blockStmt, ok := ifStmt.Consequence.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("Expected BlockStatement, got %T", ifStmt.Consequence)
	}

	if len(blockStmt.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(blockStmt.Statements))
	}

	retStmt, ok := blockStmt.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", blockStmt.Statements[0])
	}

	if retStmt.Result == nil {
		t.Fatalf("Expected return stmt to return something, got nil")
	}

	test_Identifier(t, retStmt.Result, "a")

	if ifStmt.Alternative != nil {
		t.Fatalf("Expected nil alternative, got %T", ifStmt.Alternative)
	}
}

func Test_ParseIfElseStatement(t *testing.T) {
	src := `function test() public {
        if (a > b) { return a; } else { return b; }
    }`

	file := test_helper_parseSource(t, src, false)

	fnBody := test_helper_parseFnBody(t, file)

	if len(fnBody.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(fnBody.Statements))
	}

	stmt := fnBody.Statements[0]
	ifStmt, ok := stmt.(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", stmt)
	}

	if ifStmt.Condition == nil {
		t.Fatalf("Condition expected to be <<Expression>>, got nil")
	}

	test_InfixExpression(t, ifStmt.Condition, "a", ">", "b")

	if ifStmt.Consequence == nil {
		t.Fatalf("Consequence expected to be <<BlockStatement>>, got nil")
	}

	blockStmt, ok := ifStmt.Consequence.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("Expected BlockStatement, got %T", ifStmt.Consequence)
	}

	if len(blockStmt.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(blockStmt.Statements))
	}

	retStmt, ok := blockStmt.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", blockStmt.Statements[0])
	}

	if retStmt.Result == nil {
		t.Fatalf("Expected return stmt to return something, got nil")
	}

	test_Identifier(t, retStmt.Result, "a")

	if ifStmt.Alternative == nil {
		t.Fatalf("Expected Alternative to be <<BlockStatement>>, got nil")
	}

	blockStmt, ok = ifStmt.Alternative.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("Expected BlockStatement, got %T", ifStmt.Alternative)
	}

	if len(blockStmt.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(blockStmt.Statements))
	}

	retStmt, ok = blockStmt.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", blockStmt.Statements[0])
	}

	if retStmt.Result == nil {
		t.Fatalf("Expected return stmt to return something, got nil")
	}

	test_Identifier(t, retStmt.Result, "b")
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

	if vd.Name.Value != expectedIdentifier {
		t.Errorf("Expected identifier %s, got %s",
			expectedIdentifier, vd.Name.Value)
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
