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

/*~*~*~*~*~*~*~*~*~*~*~*~* Helper Functions ~*~*~*~*~*~*~*~*~*~*~*~*~*/

func testEval(input string, boilerplate bool) object.Object {
	p := parser.Parser{}

	if boilerplate {
		input = "function test() { " + input + " }"
	}

	handle := token.NewFile("test.sol", input)
	p.Init(handle)

	file := p.ParseFile()

	return Eval(file)
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
