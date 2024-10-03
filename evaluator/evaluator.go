package evaluator

import (
	"fmt"
	"solbot/ast"
	"solbot/object"
	"solbot/token"
)

// Save commonly used object so that we are not creating new ones every time
// they show up e.g. TRUE is equal to other TRUE object, so there is no point
// in creating the new one, since the previous one is exactly the same.
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
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
		// TODO: This should probably distinguish between hex and decimal.
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
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

func nativeBoolToBooleanObject(nodeVal bool) *object.Boolean {
	if nodeVal {
		return TRUE
	}

	return FALSE
}

func evalPrefixExpression(operator token.Token, right object.Object) object.Object {
	switch operator.Type {
	default:
		return retEvalErrorObj(
			fmt.Sprintf("Unknown prefix operator: %s", operator.String()))
	case token.NOT:
		return evalNotOperatorExpression(right)
	}
}

func evalNotOperatorExpression(right object.Object) *object.Boolean {
	switch right {
	default:
		return FALSE
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	}
}

func retEvalErrorObj(message string) *object.EvalError {
	return &object.EvalError{Message: message}
}
