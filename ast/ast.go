package ast

import "solparsor/token"

type Node interface {
	Start() token.Position // First character of the node
	End() token.Position   // First character immediately after the node
}

type Expression interface {
	Node
	expressionNode()
}

type Statement interface {
	Node
	statementNode()
}

type Declaration interface {
	Node
	declarationNode()
}

type Comment struct {
	Slash token.Position // Position of the leading '/'
	Text  string
}

// In Solidity grammar it's called "SourceUnit" and represents the entire source file.
type File struct {
	Declarations []Declaration
}

func (f *File) Start() token.Position {
	if len(f.Declarations) > 0 {
		return f.Declarations[0].Start()
	}
	return 0
}

func (f *File) End() token.Position {
	if len(f.Declarations) > 0 {
		return f.Declarations[len(f.Declarations)-1].End()
	}
	return 0
}

/*~*~*~*~*~*~*~*~*~*~*~*~ Expressions *~*~*~*~*~*~*~*~*~*~*~*~*~*/

type Identifier struct {
	Position token.Position
	Name     string
}

// In Solidity grammar: "ElementaryTypeName"
// address, address payable, bool, string, uint, int, bytes, fixed, fixed-bytes and ufixed
type ElementaryType struct {
	ValuePos token.Position
	Kind     token.Token
	Value    string
}

/*~*~*~*~*~*~*~*~*~*~*~*~* Statements *~*~*~*~*~*~*~*~*~*~*~*~*~*/

/*~*~*~*~*~*~*~*~*~*~*~*~ Declarations ~*~*~*~*~*~*~*~*~*~*~*~*~*/
