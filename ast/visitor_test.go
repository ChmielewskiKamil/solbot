package ast_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ChmielewskiKamil/solbot/ast"
	"github.com/ChmielewskiKamil/solbot/parser"
)

type mockVisitor struct {
	visited []ast.Node
}

func (v *mockVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		v.visited = append(v.visited, node)
	}
	return v
}

func TestWalk(t *testing.T) {
	source := `
contract Simple {
    function add(uint256 a, uint256) {
        return a + c;
    }
}`
	astRoot, err := parser.ParseFile("test.sol", strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	expectedTypes := []any{
		(*ast.File)(nil),
		(*ast.ContractDeclaration)(nil),
		(*ast.Identifier)(nil), // Contract name 'Simple'
		(*ast.ContractBody)(nil),
		(*ast.FunctionDeclaration)(nil), // Function keyword
		(*ast.Identifier)(nil),          // Function name
		(*ast.ParamList)(nil),
		(*ast.Param)(nil),
		(*ast.ElementaryType)(nil), // uint256
		(*ast.Identifier)(nil),     // 'a'
		(*ast.Param)(nil),
		(*ast.ElementaryType)(nil), // uint256 (identifier omitted)
		(*ast.BlockStatement)(nil),
		(*ast.ReturnStatement)(nil),
		(*ast.InfixExpression)(nil),
		(*ast.Identifier)(nil), // 'a'
		(*ast.Identifier)(nil), // 'c'
	}

	visitor := &mockVisitor{}
	ast.Walk(visitor, astRoot)

	if len(visitor.visited) != len(expectedTypes) {
		t.Fatalf("Expected to visit %d nodes, but visited %d",
			len(expectedTypes), len(visitor.visited))
	}

	for i, expectedType := range expectedTypes {
		visitedNode := visitor.visited[i]

		visitedType := reflect.TypeOf(visitedNode)
		expectedReflectType := reflect.TypeOf(expectedType)

		if visitedType != expectedReflectType {
			t.Fatalf("Node %d: incorrect type. \n- want: %v\n- got:  %v",
				i, expectedReflectType, visitedType)
		}
	}
}
