package evaluator

import (
	"solbot/ast"
	"solbot/object"
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// File

	case *ast.File:
		return evalDeclarations(node.Declarations)

		// Declarations

	case *ast.FunctionDeclaration:
		// TODO: Hacky way to eval other parts in tests, fix later.
		return Eval(node.Body)

	// Statements

	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BlockStatement:
		return evalStatements(node.Statements)

	// Expressions

	case *ast.NumberLiteral:
		return &object.Integer{Value: node.Value}
	}

	return nil
}

func evalDeclarations(decls []ast.Declaration) object.Object {
	var result object.Object

	for _, decl := range decls {
		result = Eval(decl)
	}

	return result
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt)
	}

	return result
}
