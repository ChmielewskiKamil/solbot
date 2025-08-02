package ast

// Visitor defines the interface for an AST visitor.
// The Visit method is called for each node encountered by Walk.
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk traverses an AST in depth-first order.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *File:
		for _, decl := range n.Declarations {
			Walk(v, decl)
		}

	case *ContractDeclaration:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		for _, parent := range n.Parents {
			if parent != nil {
				Walk(v, parent)
			}
		}

		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *ContractBody:
		for _, decl := range n.Declarations {
			Walk(v, decl)
		}

	case *FunctionDeclaration:
		if n.Name != nil {
			Walk(v, n.Name)
		}

		if n.Params != nil {
			Walk(v, n.Params)
		}

		if n.Results != nil {
			Walk(v, n.Results)
		}

		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *ParamList:
		for _, param := range n.List {
			if param != nil {
				Walk(v, param)
			}
		}

	case *Param:
		if n.Type != nil {
			Walk(v, n.Type)
		}

		if n.Name != nil {
			Walk(v, n.Name)
		}

		// Data Location is not an ast.Node, so it is skipped.
		// It is an attribute of the Node.

	case *BlockStatement:
		for _, stmt := range n.Statements {
			Walk(v, stmt)
		}

	case *ReturnStatement:
		if n.Result != nil {
			Walk(v, n.Result)
		}

	case *InfixExpression:
		if n.Left != nil {
			Walk(v, n.Left)
		}

		if n.Right != nil {
			Walk(v, n.Right)
		}

	////// Leaf Node Cases //////
	// These nodes have no children, so their cases are empty,
	// but they must be present in the switch so that API consumer know
	// when these were visited.
	case
		*Identifier,
		*ElementaryType:
		// No children to walk.
	}

	v.Visit(nil)
}
