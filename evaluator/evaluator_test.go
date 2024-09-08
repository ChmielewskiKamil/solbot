package evaluator

import (
	"math/big"
	"solbot/object"
	"solbot/parser"
	"solbot/token"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
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
