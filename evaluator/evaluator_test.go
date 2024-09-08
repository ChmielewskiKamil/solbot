package evaluator

import (
	"math/big"
	"solbot/object"
	"solbot/parser"
	"solbot/token"
	"testing"
)

func Test_EvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected *big.Int
	}{
		{"1", big.NewInt(1)},
		{"50", big.NewInt(50)},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func Test_EvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input, true)
		testBooleanObject(t, evaluated, tt.expected)
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
