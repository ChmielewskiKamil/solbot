package evaluator

import (
	"math/big"
	"solbot/object"
	"solbot/parser"
	"solbot/token"
	"testing"
)

func Test_Eval_IntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected *big.Int
	}{
		{"1", big.NewInt(1)},
		{"50", big.NewInt(50)},
		{"-1", big.NewInt(-1)},
		{"-50", big.NewInt(-50)},
		{"-2 + -2 + 4", big.NewInt(0)},
		{"-10 + 10 - 10", big.NewInt(-10)},
		{"10 + 5 * 2 - 10 / 2", big.NewInt(15)},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", big.NewInt(50)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func Test_Eval_BooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 > 2", false},
		{"2 > 1", true},
		{"2 < 3", true},
		{"2 < 1", false},
		{"3 == 3", true},
		{"3 == 5", false},
		{"4 != 4", false},
		{"5 != 5", false},
		{"true == true", true},
		{"false != true", true},
		{"true != true", false},
		{"true == false", false},
		{"(2 > 5) == false", true},
		{"(5 > 2) == true", true},
		{"false != (5 > 2)", true},
		{"true != (2 > 5)", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

// token.NOT is ! (bang)
func Test_Eval_NotOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func Test_Eval_IfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", big.NewInt(10)},
		{"if (false) { 10 }", nil},
		{"if (1 < 2) { 10 }", big.NewInt(10)},
		{"if (5 > 3) { 10 }", big.NewInt(10)},
		{"if (5 < 3) { 10 }", nil},
		{"if (10 > 20) { 10 } else { 20 }", big.NewInt(20)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		expectedInt, ok := tt.expected.(*big.Int)
		if ok {
			testIntegerObject(t, evaluated, expectedInt)
		} else {
			if evaluated != nil {
				t.Errorf("Expected nil, got: %T", evaluated)
			}
		}
	}
}

func Test_Eval_ReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected *big.Int
	}{
		{"return 5;", big.NewInt(5)},
		{"1 + 2; return 5", big.NewInt(5)},
		{"1 + 2; return 5; 5 * 3;", big.NewInt(5)},
		{`if (2 > 1) { 
            if (3 > 2) { 
                return 3; 
            } return 5; 
        }`, big.NewInt(3)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func Test_Eval_ErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedMsg string
	}{
		{"1 + true", "Incorrect object types for infix expression: INTEGER + BOOLEAN."},
		{"-true", "The '-' prefix operator can only be used with integers. Got: BOOLEAN instead."},
		{"!5", "The '!' prefix operator can only be used with booleans. Got: INTEGER instead."},
		{"1 >> 2", "Incorrect operator: '>>' in integer infix expression."},
		{`if (5 > 3) {
            true + 5;
        }`, "Incorrect object types for infix expression: BOOLEAN + INTEGER."},
		{`if (5 > 3) {
            if (10 > 5) {
                true * false;
            }
        }`, "Incorrect object types for infix expression: BOOLEAN * BOOLEAN."},
		{"foo", "Identifier not found: foo"},
		// {"", ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		errObj, ok := evaluated.(*object.EvalError)
		if !ok {
			t.Errorf("Expected eval error object, got: %T(%+v)",
				evaluated, evaluated)
			// If we didn't get the err obj we can't access its message.
			continue
		}

		if errObj.Message != tt.expectedMsg {
			t.Errorf("Err message is wrong. Expected: %q, got: %q",
				tt.expectedMsg, errObj.Message)
		}
	}
}

func Test_Eval_VariableDeclarationStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected *big.Int
	}{
		{"uint256 a = 5; uint256 b = 5; return a + b;", big.NewInt(10)},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func Test_Eval_FunctionDeclaration(t *testing.T) {
	input := `function add(uint256 a, uint256 b) public virtual pure returns (uint256) {
        return a + b;
    }`

	evaluated := testEval(input, false)

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("Object is not a Function. got=%T (%+v)", evaluated, evaluated)
	}

	if fn.Name.String() != "add" {
		t.Fatalf("Function's name is not add. got=%s", fn.Name.String())
	}

	if len(fn.Params.List) != 2 {
		t.Fatalf("Function has wrong parameterst. got=%+v", fn.Params.List)
	}

	if fn.Params.List[0].Name.String() != "a" {
		t.Fatalf("Function param incorrect. got=%s", fn.Params.List[0].Name.String())
	}
}

/*~*~*~*~*~*~*~*~*~*~*~*~* Helper Functions ~*~*~*~*~*~*~*~*~*~*~*~*~*/

func testEval(input string, boilerplate bool) object.Object {
	p := parser.Parser{}
	env := object.NewEnvironment()

	if boilerplate {
		input = "function test() { " + input + " }"
	}

	handle := token.NewFile("test.sol", input)
	p.Init(handle)

	file := p.ParseFile()

	evaluated := Eval(file, env)

	if boilerplate {
		fn, ok := evaluated.(*object.Function)
		if !ok {
			panic("Test boilerplate does not work, should have wrapped in a function.")
		}

		result := Eval(fn.Body, env)
		if retValue, ok := result.(*object.ReturnValue); ok {
			return retValue.Value
		}
		return result
	}

	return evaluated
}

func testIntegerObject(t *testing.T, obj object.Object, expected *big.Int) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("Expected object.Integer, got %T (%+v)", obj, obj)
		return false
	}

	if result.Value.Cmp(expected) != 0 {
		t.Errorf("Expected %s, got %s", expected.String(), result.Value.String())
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("Expected object.Boolean, got %T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("Expected %t, got %t", expected, result.Value)
		return false
	}

	return true
}
