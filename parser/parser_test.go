package parser_test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ChmielewskiKamil/solbot/ast"
	"github.com/ChmielewskiKamil/solbot/parser"
	"github.com/ChmielewskiKamil/solbot/token"
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
	// TODO: When overriding the param identifier can be empty?
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

		// TODO: return address(0) is not tested because it is an elementary type.

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

func Test_ParseEventDeclaration(t *testing.T) {
	src := `contract Test {
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address spender, uint256 amount);
    event Log(string message) anonymous;
    event ComplexEvent(uint256 indexed id, bool flag, bytes32 data);
    }
    `

	file := test_helper_parseSource(t, src, false)

	if len(file.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(file.Declarations))
	}

	decl := file.Declarations[0]
	contract, ok := decl.(*ast.ContractDeclaration)
	if !ok {
		t.Fatalf("Expected contract declaration, got: %T", decl)
	}

	tests := []struct {
		expectedName   string
		isAnonymous    bool
		expectedParams []*struct {
			name      string
			typeToken token.TokenType
			isIndexed bool
		}
	}{
		{
			expectedName: "Transfer",
			isAnonymous:  false,
			expectedParams: []*struct {
				name      string
				typeToken token.TokenType
				isIndexed bool
			}{
				{"from", token.ADDRESS, true},
				{"to", token.ADDRESS, true},
				{"value", token.UINT_256, false},
			},
		},
		{
			expectedName: "Approval",
			isAnonymous:  false,
			expectedParams: []*struct {
				name      string
				typeToken token.TokenType
				isIndexed bool
			}{
				{"owner", token.ADDRESS, true},
				{"spender", token.ADDRESS, false},
				{"amount", token.UINT_256, false},
			},
		},
		{
			expectedName: "Log",
			isAnonymous:  true,
			expectedParams: []*struct {
				name      string
				typeToken token.TokenType
				isIndexed bool
			}{
				{"message", token.STRING, false},
			},
		},
		{
			expectedName: "ComplexEvent",
			isAnonymous:  false,
			expectedParams: []*struct {
				name      string
				typeToken token.TokenType
				isIndexed bool
			}{
				{"id", token.UINT_256, true},
				{"flag", token.BOOL, false},
				{"data", token.BYTES_32, false},
			},
		},
	}

	for i, tt := range tests {
		eventDecl, ok := contract.Body.Declarations[i].(*ast.EventDeclaration)
		if !ok {
			t.Errorf("Test %d: Expected EventDeclaration, got %T", i, contract.Body.Declarations[i])
			continue
		}

		// Check event name
		if eventDecl.Name.String() != tt.expectedName {
			t.Errorf("Test %d: Expected name %s, got %s", i, tt.expectedName, eventDecl.Name.String())
		}

		// Check anonymity
		if eventDecl.IsAnonymous != tt.isAnonymous {
			t.Errorf("Test %d: Expected isAnonymous %v, got %v", i, tt.isAnonymous, eventDecl.IsAnonymous)
		}

		// Check parameters
		if len(eventDecl.Params.List) != len(tt.expectedParams) {
			t.Errorf("Test %d: Expected %d parameters, got %d", i, len(tt.expectedParams), len(eventDecl.Params.List))
			continue
		}

		for j, paramTest := range tt.expectedParams {
			param := eventDecl.Params.List[j]
			if param.Name.String() != paramTest.name {
				t.Errorf("Test %d, Param %d: Expected name %s, got %s", i, j, paramTest.name, param.Name.String())
			}
			if param.Type.(*ast.ElementaryType).Kind.Type != paramTest.typeToken {
				t.Errorf("Test %d, Param %d: Expected type %v, got %v", i, j, paramTest.typeToken, param.Type.(*ast.ElementaryType).Kind.Type)
			}
			if param.IsIndexed != paramTest.isIndexed {
				t.Errorf("Test %d, Param %d: Expected isIndexed %v, got %v", i, j, paramTest.isIndexed, param.IsIndexed)
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

func TestParseDeclarations(t *testing.T) {
	testCases := []struct {
		name     string
		source   string
		validate func(t *testing.T, decls []ast.Declaration)
	}{
		{
			name: "contract with state variables",
			source: `
		// Comment here should be ignored by parser.
                contract Test {
                    uint256 balance = 100;
		    // Comment here should be ignored by parser.
                    bool constant IS_OWNER = true;
                }
            `,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				contract, ok := decls[0].(*ast.ContractDeclaration)
				if !ok {
					t.Fatalf("Expected ContractDeclaration, got %T", decls[0])
				}
				if len(contract.Body.Declarations) != 2 {
					t.Fatalf("Expected 2 state variables, got %d", len(contract.Body.Declarations))
				}
			},
		},
		{
			name: "simple function declaration",
			source: `
                function getBalance() public view returns (uint256) {
		    // Comment here should be ignored by parser.
                    return 10;
                }
            `,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				fn, ok := decls[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Fatalf("Expected FunctionDeclaration, got %T", decls[0])
				}
				if fn.Name.Value != "getBalance" {
					t.Errorf("Expected function name 'getBalance', got '%s'", fn.Name.Value)
				}
			},
		},
		{
			name: "Counter contract with comments",
			source: `
                contract Counter {
                    uint256 public count;
                    // Function to decrement count by 1
                    function dec() public {
                        // This function will fail if count = 0
                        count -= 1;
                    }
                }
            `,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				contract, ok := decls[0].(*ast.ContractDeclaration)
				if !ok {
					t.Fatalf("Expected ContractDeclaration, got %T", decls[0])
				}
				// The parser should find 2 declarations inside the contract:
				// the state variable and the function. The comments are ignored.
				if len(contract.Body.Declarations) != 2 {
					t.Fatalf("Expected 2 declarations in contract body, got %d", len(contract.Body.Declarations))
				}

				fn, ok := contract.Body.Declarations[1].(*ast.FunctionDeclaration)
				if !ok {
					t.Fatalf("Expected second declaration to be a Function, got %T", contract.Body.Declarations[1])
				}

				if len(fn.Body.Statements) != 1 {
					t.Fatalf("Expected 1 statement in function body, got %d", len(fn.Body.Statements))
				}
			},
		}, {
			name: "Tuple: Variable Declarations no ommissions",
			source: `function _() {
			(address owner, uint256 balance) = getValues();
		}`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				fn, ok := decls[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Fatalf("Expected FunctionDeclaration, got %T", decls[0])
				}
				if len(fn.Body.Statements) != 1 {
					t.Fatalf("Expected 1 statement in function, got %d", len(fn.Body.Statements))
				}
				vdTupleStmt, ok := fn.Body.Statements[0].(*ast.VariableDeclarationTupleStatement)
				if !ok {
					t.Fatalf("Expected variable declaration tuple statement, got %T, ", vdTupleStmt)
				}
				if len(vdTupleStmt.Declarations) != 2 {
					t.Fatalf("Expected 2 variable declarations inside a tuple, got %d", len(vdTupleStmt.Declarations))
				}
			},
		}, {
			name: "Tuple: Variable Declarations with ommissions",
			source: `function _() {
			(address owner,,,uint256 balance) = getValues();
		}`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				fn, ok := decls[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Fatalf("Expected FunctionDeclaration, got %T", decls[0])
				}
				if len(fn.Body.Statements) != 1 {
					t.Fatalf("Expected 1 statement in function, got %d", len(fn.Body.Statements))
				}
				vdTupleStmt, ok := fn.Body.Statements[0].(*ast.VariableDeclarationTupleStatement)
				if !ok {
					t.Fatalf("Expected variable declaration tuple statement, got %T, ", vdTupleStmt)
				}
				if len(vdTupleStmt.Declarations) != 4 {
					t.Fatalf("Expected 4 variable declarations inside a tuple, got %d", len(vdTupleStmt.Declarations))
				}
			},
		}, {
			name: "Tuple: Variable Declarations with leading ommissions",
			source: `function _() {
			(,,,uint256 balance) = getValues();
		}`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				fn, ok := decls[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Fatalf("Expected FunctionDeclaration, got %T", decls[0])
				}
				if len(fn.Body.Statements) != 1 {
					t.Fatalf("Expected 1 statement in function, got %d", len(fn.Body.Statements))
				}
				vdTupleStmt, ok := fn.Body.Statements[0].(*ast.VariableDeclarationTupleStatement)
				if !ok {
					t.Fatalf("Expected variable declaration tuple statement, got %T, ", vdTupleStmt)
				}
				if len(vdTupleStmt.Declarations) != 4 {
					t.Fatalf("Expected 4 variable declarations inside a tuple, got %d", len(vdTupleStmt.Declarations))
				}
			},
		},
		{
			name:   "Tuple: Complex declarations with various empty slots",
			source: `function _() { (uint256 a, , bytes memory c, ,) = getValues(); }`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 function declaration, got %d", len(decls))
				}
				fn, ok := decls[0].(*ast.FunctionDeclaration)
				if !ok {
					t.Fatalf("Expected FunctionDeclaration, got %T", decls[0])
				}
				if len(fn.Body.Statements) != 1 {
					t.Fatalf("Expected 1 statement in function, got %d", len(fn.Body.Statements))
				}
				vdTupleStmt, ok := fn.Body.Statements[0].(*ast.VariableDeclarationTupleStatement)
				if !ok {
					t.Fatalf("Expected VariableDeclarationTupleStatement, got %T", fn.Body.Statements[0])
				}

				// Define the expected structure for each slot in the tuple.
				// A 'nil' entry represents an expected empty slot.
				type expectedPart struct {
					Type         token.TokenType
					Name         string
					DataLocation ast.DataLocation
				}

				expectedDecls := []*expectedPart{
					{Type: token.UINT_256, Name: "a", DataLocation: ast.NO_DATA_LOCATION},
					nil,
					{Type: token.BYTES, Name: "c", DataLocation: ast.Memory},
					nil,
					nil,
				}

				if len(vdTupleStmt.Declarations) != len(expectedDecls) {
					t.Fatalf("Expected %d declarations in tuple, got %d", len(expectedDecls), len(vdTupleStmt.Declarations))
				}

				// Now, iterate and check each declaration part.
				for i, expected := range expectedDecls {
					actual := vdTupleStmt.Declarations[i]

					// Case 1: We expect an empty slot.
					if expected == nil {
						if actual != nil {
							t.Errorf("Test %d: Expected a nil declaration (empty slot), but got a non-nil one.", i)
						}
						continue // Slot is correctly empty, move to the next.
					}

					// Case 2: We expect a full declaration.
					if actual == nil {
						t.Errorf("Test %d: Expected declaration for '%s', but got a nil (empty slot).", i, expected.Name)
						continue // Can't test further, move to the next.
					}

					// Check the name
					if actual.Name.Value != expected.Name {
						t.Errorf("Test %d: Expected name '%s', got '%s'", i, expected.Name, actual.Name.Value)
					}

					// Check the data location
					if actual.DataLocation != expected.DataLocation {
						t.Errorf("Test %d: Expected data location %s for '%s', got %s", i, expected.DataLocation, expected.Name, actual.DataLocation)
					}

					// Check the type
					elemType, ok := actual.Type.(*ast.ElementaryType)
					if !ok {
						t.Errorf("Test %d: Expected ElementaryType, got %T", i, actual.Type)
						continue
					}
					if elemType.Kind.Type != expected.Type {
						t.Errorf("Test %d: Expected type %s, got %s", i, expected.Type, elemType.Kind.Type)
					}
				}
			},
		},
		{
			name: "using for simple library binding",
			source: `
		using SafeMath for uint256;
	`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}

				dir, ok := decls[0].(*ast.UsingForDirective)
				if !ok {
					t.Fatalf("Expected UsingForDirective, got %T", decls[0])
				}

				if dir.LibraryName == nil || dir.LibraryName.Value != "SafeMath" {
					t.Errorf("Expected library name 'SafeMath', got %v", dir.LibraryName)
				}
				if dir.List != nil {
					t.Errorf("Expected List to be nil, got %v", dir.List)
				}
				if dir.IsWildcard {
					t.Error("Expected IsWildcard to be false")
				}
				if dir.IsGlobal {
					t.Error("Expected IsGlobal to be false")
				}

				// Check the type
				ft, ok := dir.ForType.(*ast.ElementaryType)
				if !ok {
					t.Fatalf("Expected ForType to be ElementaryType, got %T", dir.ForType)
				}
				if ft.Kind.Type != token.UINT_256 {
					t.Errorf("Expected ForType to be uint256, got %s", ft.Kind.Literal)
				}
			},
		},
		{
			name: "using for wildcard binding in contract",
			source: `
		contract C {
			using Lib for *;
		}
	`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 contract declaration, got %d", len(decls))
				}
				contract, ok := decls[0].(*ast.ContractDeclaration)
				if !ok {
					t.Fatalf("Expected ContractDeclaration, got %T", decls[0])
				}
				if len(contract.Body.Declarations) != 1 {
					t.Fatalf("Expected 1 declaration in contract, got %d", len(contract.Body.Declarations))
				}

				dir, ok := contract.Body.Declarations[0].(*ast.UsingForDirective)
				if !ok {
					t.Fatalf("Expected UsingForDirective, got %T", contract.Body.Declarations[0])
				}

				if dir.LibraryName == nil || dir.LibraryName.Value != "Lib" {
					t.Errorf("Expected library name 'Lib', got %v", dir.LibraryName)
				}
				if !dir.IsWildcard {
					t.Error("Expected IsWildcard to be true")
				}
				if dir.ForType != nil {
					t.Errorf("Expected ForType to be nil for wildcard, got %v", dir.ForType)
				}
			},
		},
		{
			name: "using for list binding with global scope",
			source: `
		using {add, sub} for uint256 global;
	`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				dir, ok := decls[0].(*ast.UsingForDirective)
				if !ok {
					t.Fatalf("Expected UsingForDirective, got %T", decls[0])
				}

				if dir.LibraryName != nil {
					t.Errorf("Expected LibraryName to be nil, got %v", dir.LibraryName)
				}
				if len(dir.List) != 2 {
					t.Fatalf("Expected 2 items in List, got %d", len(dir.List))
				}

				// Check first item
				if dir.List[0].Path.Value != "add" {
					t.Errorf("Expected item 0 path 'add', got '%s'", dir.List[0].Path.Value)
				}
				if dir.List[0].Alias.Type != token.ILLEGAL {
					t.Errorf("Expected item 0 to have no alias, got %v", dir.List[0].Alias)
				}

				// Check second item
				if dir.List[1].Path.Value != "sub" {
					t.Errorf("Expected item 1 path 'sub', got '%s'", dir.List[1].Path.Value)
				}
				if dir.List[1].Alias.Type != token.ILLEGAL {
					t.Errorf("Expected item 1 to have no alias, got %v", dir.List[1].Alias)
				}

				if !dir.IsGlobal {
					t.Error("Expected IsGlobal to be true")
				}
			},
		},
		{
			name: "using for complex list with identifier and operator aliases",
			source: `
		using {add as +, isEqual as ==, sub} for MyType global;
	`,
			validate: func(t *testing.T, decls []ast.Declaration) {
				if len(decls) != 1 {
					t.Fatalf("Expected 1 declaration, got %d", len(decls))
				}
				dir, ok := decls[0].(*ast.UsingForDirective)
				if !ok {
					t.Fatalf("Expected UsingForDirective, got %T", decls[0])
				}

				if len(dir.List) != 3 {
					t.Fatalf("Expected 3 items in List, got %d", len(dir.List))
				}

				expected := []struct {
					path      string
					aliasType token.TokenType
					aliasLit  string
				}{
					{"add", token.ADD, "+"},
					{"isEqual", token.EQUAL, "=="},
					{"sub", token.ILLEGAL, ""},
				}

				for i, item := range dir.List {
					if item.Path.Value != expected[i].path {
						t.Errorf("Item %d: Expected path '%s', got '%s'", i, expected[i].path, item.Path.Value)
					}
					if item.Alias.Type != expected[i].aliasType {
						t.Errorf("Item %d: Expected alias type %s, got %s", i, expected[i].aliasType, item.Alias.Type)
					}
					if item.Alias.Literal != expected[i].aliasLit {
						t.Errorf("Item %d: Expected alias literal '%s', got '%s'", i, expected[i].aliasLit, item.Alias.Literal)
					}
				}

				if !dir.IsGlobal {
					t.Error("Expected IsGlobal to be true")
				}

				ft, ok := dir.ForType.(*ast.UserDefinedType)
				if !ok {
					t.Fatalf("Expected ForType to be Identifier (for user-defined type), got %T", dir.ForType)
				}
				if ft.Name.Value != "MyType" {
					t.Errorf("Expected ForType to be 'MyType', got '%s'", ft.Name.Value)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := parser.ParseFile("test.sol", strings.NewReader(tc.source))
			if err != nil {
				t.Fatalf("ParseFile failed: %v", err)
			}
			if file == nil {
				t.Fatalf("ParseFile returned a nil file")
			}

			tc.validate(t, file.Declarations)
		})
	}
}
