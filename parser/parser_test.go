package parser

import (
	"math/big"
	"solbot/ast"
	"solbot/token"
	"testing"
)

func Test_ParseStateVariableDeclaration(t *testing.T) {
	src := `contract Test {
    address owner = 0x12345;
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
    }
    `

	file := test_helper_parseSource(t, src, false)

	if len(file.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(file.Declarations))
	}

	tests := []struct {
		expectedType       token.TokenType
		expectedMutability ast.Mutability
		expectedVisibility ast.Visibility
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{token.ADDRESS, 0, ast.Internal, "owner", big.NewInt(0x12345)},
		{token.UINT_256, 0, ast.Internal, "balance", big.NewInt(100)},
		{token.BOOL, 0, ast.Internal, "isOwner", true},
		{token.BOOL, 4, ast.Internal, "IS_OWNER", true},
		{token.BOOL, 4, ast.Internal, "isOwner", false},
		{token.BOOL, 4, ast.Internal, "is_owner", false},
		{token.UINT_256, 0, ast.Internal, "balance", big.NewInt(100)},
		{token.ADDRESS, 4, ast.Internal, "router", big.NewInt(0x1337)},
		{token.BOOL, 0, ast.Internal, "isOwner", true},
		{token.UINT_16, 4, ast.Internal, "ONE_hundred_IS_100", big.NewInt(100)},
		{token.UINT_256, 4, ast.Internal, "DENOMINATOR", big.NewInt(1_000_000)},
		{token.UINT_256, 4, ast.Private, "Is_This_Snake_Case", big.NewInt(0)},
		{token.UINT_256, 6, ast.Internal, "blob", big.NewInt(0)},
	}

	decl := file.Declarations[0]
	contract, ok := decl.(*ast.ContractDeclaration)
	if !ok {
		t.Fatalf("Expected contract declaration, got: %T", decl)
	}

	for i, tt := range tests {
		stateVar := contract.Body.Declarations[i]
		if !testParseElementaryType(
			t,
			stateVar,
			tt.expectedType,
			tt.expectedMutability,
			tt.expectedVisibility,
			tt.expectedIdentifier,
			tt.expectedValue) {
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
        return tester;
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
		expectedTkType     token.TokenType
		expectedIdentifier string
	}{
		{token.ADDRESS, "owner"},
		{token.UINT_256, "amount"},
	}

	for i, tt := range tests {
		param := fd.Params.List[i]
		if param.Type == nil {
			t.Fatalf("Expected ElementaryType, got nil")
		}

		et, ok := param.Type.(*ast.ElementaryType)
		if !ok {
			t.Fatalf("Expected ElementaryType, got %T", param.Type)
		}

		if et.Kind.Type != tt.expectedTkType {
			t.Errorf("Expected token type %T, got %T", tt.expectedTkType, et.Kind.Type)
		}

		test_LiteralExpression(t, param.Name, tt.expectedIdentifier)
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
	rtStmt, ok := stmt.(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", stmt)
	}

	if rtStmt.Result == nil {
		t.Fatalf("Expected return stmt to return something, got nil")
	}

	test_Identifier(t, rtStmt.Result, "tester")
}

func Test_ParseContractDeclaration(t *testing.T) {
	src := `
    contract MyContract is BaseContract {
        uint256 public myVar;

        function myFunction() public view returns (uint256) {
            uint256 balance;
            return myVar;
        }

        function test() public {}
    }
    `

	file := test_helper_parseSource(t, src, false)

	if len(file.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(file.Declarations))
	}

	decl := file.Declarations[0]
	contract, ok := decl.(*ast.ContractDeclaration)
	if !ok {
		t.Fatalf("Expected ContractDeclaration, got %T", decl)
	}

	// Verify the contract's name
	if contract.Name.Value != "MyContract" {
		t.Errorf("Expected contract name MyContract, got %s", contract.Name.Value)
	}

	// Verify inheritance
	if len(contract.Parents) != 1 {
		t.Fatalf("Expected 1 parent contract, got %d", len(contract.Parents))
	}

	if contract.Parents[0].Value != "BaseContract" {
		t.Errorf("Expected parent contract BaseContract, got %s", contract.Parents[0].Value)
	}

	// Verify the body
	body := contract.Body
	if body == nil {
		t.Fatalf("Expected ContractBody, got nil")
	}

	if len(body.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations in the contract body, got %d", len(body.Declarations))
	}

	// Verify state variable
	varDecl, ok := body.Declarations[0].(*ast.StateVariableDeclaration)
	if !ok {
		t.Fatalf("Expected VariableDeclaration, got %T", body.Declarations[0])
	}

	if varDecl.Name.Value != "myVar" {
		t.Errorf("Expected variable name myVar, got %s", varDecl.Name.Value)
	}

	// Verify function declaration
	funcDecl, ok := body.Declarations[1].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("Expected FunctionDeclaration, got %T", body.Declarations[1])
	}

	if funcDecl.Name.Value != "myFunction" {
		t.Errorf("Expected function name myFunction, got %s", funcDecl.Name.Value)
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

	tests := []struct {
		expectedValue interface{}
	}{
		{big.NewInt(10)},
		{big.NewInt(0x12345)},
		{true},
		{"staked"},
	}

	for i, tt := range tests {
		stmt := fnBody.Statements[i]
		retStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("Expected ReturnStatement, got %T", stmt)
		}

		if retStmt.Result == nil {
			t.Fatalf("Expected return stmt to return something, got nil")
		}

		// @TODO: return address(0) is not tested because it is an elementary type.

		test_LiteralExpression(t, retStmt.Result, tt.expectedValue)
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

func Test_ParseParameterList(t *testing.T) {
	tests := []struct {
		src            string
		expectedIdents []string
		expectedTypes  []token.TokenType
	}{
		{`function test() {}`, []string{}, []token.TokenType{}},
		{`function test(uint256 a) {}`, []string{"a"}, []token.TokenType{token.UINT_256}},
		{`function test(uint256 a, bool b) {}`, []string{"a", "b"}, []token.TokenType{token.UINT_256, token.BOOL}},
	}

	for _, tt := range tests {
		file := test_helper_parseSource(t, tt.src, false)

		decl := file.Declarations[0]
		fd, ok := decl.(*ast.FunctionDeclaration)
		if !ok {
			t.Fatalf("Expected FunctionDeclaration, got %T", decl)
		}

		if fd.Params == nil {
			t.Fatalf("Expected ParamList, got nil")
		}

		if len(fd.Params.List) != len(tt.expectedIdents) {
			t.Fatalf("Expected %d parameters, got %d", len(tt.expectedIdents), len(fd.Params.List))
		}

		for i, param := range fd.Params.List {
			test_LiteralExpression(t, param.Name, tt.expectedIdents[i])

			et, ok := param.Type.(*ast.ElementaryType)
			if !ok {
				t.Fatalf("Expected ElementaryType, got %T", param.Type)
			}

			if et.Kind.Type != tt.expectedTypes[i] {
				t.Errorf("Expected token type %T, got %T", tt.expectedTypes[i], et.Kind.Type)
			}
		}

	}
}

func testParseElementaryType(t *testing.T, decl ast.Declaration,
	expectedType token.TokenType, expectedMutability ast.Mutability,
	expectedVisibility ast.Visibility, expectedIdentifier string,
	expectedValue interface{}) bool {
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

	if vd.Value == nil {
		t.Fatalf("Expected value, got nil")
	}

	test_LiteralExpression(t, vd.Value, expectedValue)

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
