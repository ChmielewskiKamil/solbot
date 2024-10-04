package evaluator

import (
	"fmt"
	"math/big"
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
	default:
		return retEvalErrorObj(fmt.Sprintf("Unhandled ast node: %T", node))

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
	case *ast.IfStatement:
		return evalIfStatement(node)

	// Expressions

	case *ast.NumberLiteral:
		// TODO: This should probably distinguish between hex and decimal.
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.ReturnStatement:
		result := Eval(node.Result)
		return &object.ReturnValue{Value: result}
	}
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

		// If we encounter a return statement, we have to return earlier.
		if retValue, ok := result.(*object.ReturnValue); ok {
			return retValue.Value
		}
	}

	return result
}

func evalPrefixExpression(operator token.Token, right object.Object) object.Object {
	switch operator.Type {
	default:
		return retEvalErrorObj(
			fmt.Sprintf("Unknown prefix operator: %s", operator.String()))
	case token.NOT:
		return evalNotPrefixOperatorExpression(right)
	case token.SUB:
		return evalSubPrefixOperatorExpression(right)
	}
}

func evalNotPrefixOperatorExpression(right object.Object) object.Object {
	switch right {
	default:
		// TODO: Why do we return true in the default case?
		return FALSE
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	}
}

func evalSubPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return retEvalErrorObj(
			fmt.Sprintf(
				"The '-' prefix operator can only be used with integers. Got: %T instead.",
				right))
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: *new(big.Int).Neg(&value)}
}

func evalInfixExpression(
	operator token.Token,
	left object.Object,
	right object.Object) object.Object {

	switch {
	default:
		return retEvalErrorObj("Incorrect object types for infix expression.")
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator.Type == token.EQUAL:
		// Comparing two objects uses pointer comparison. Since boolean objects
		// TRUE and FALSE are always the same (point to the same memory address)
		// we can compare them right away here. For integer objects we allocate
		// new object each time and have to compare the value stored inside them.
		// For integers 5 == 5 when compared on objects would return false.
		return nativeBoolToBooleanObject(left == right)
	case operator.Type == token.NOT_EQUAL:
		return nativeBoolToBooleanObject(left != right)
	}
}

func evalIntegerInfixExpression(
	operator token.Token,
	left object.Object,
	right object.Object) object.Object {

	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator.Type {
	default:
		return retEvalErrorObj(
			fmt.Sprintf("Unhandled: %s operator when evaluating infix expression.",
				operator.String()))
	case token.ADD:
		result := new(big.Int).Add(&leftVal, &rightVal)
		return &object.Integer{Value: *result}
	case token.SUB:
		result := new(big.Int).Sub(&leftVal, &rightVal)
		return &object.Integer{Value: *result}
	case token.MUL:
		result := new(big.Int).Mul(&leftVal, &rightVal)
		return &object.Integer{Value: *result}
	case token.DIV:
		result := new(big.Int).Div(&leftVal, &rightVal)
		return &object.Integer{Value: *result}
	case token.GREATER_THAN:
		isGreater := leftVal.Cmp(&rightVal) == 1
		return nativeBoolToBooleanObject(isGreater)
	case token.LESS_THAN:
		isLess := leftVal.Cmp(&rightVal) == -1
		return nativeBoolToBooleanObject(isLess)
	case token.EQUAL:
		isEqual := leftVal.Cmp(&rightVal) == 0
		return nativeBoolToBooleanObject(isEqual)
	case token.NOT_EQUAL:
		isNotEqual := leftVal.Cmp(&rightVal) != 0
		return nativeBoolToBooleanObject(isNotEqual)
	}
}

func evalIfStatement(ifStmt *ast.IfStatement) object.Object {
	evaluated := Eval(ifStmt.Condition)
	cond, ok := evaluated.(*object.Boolean)
	if !ok {
		return retEvalErrorObj(fmt.Sprintf(
			"The condition has to be an *object.Boolean, got: %T instead.",
			evaluated))
	}

	if cond.Value == true {
		return Eval(ifStmt.Consequence)
	} else if ifStmt.Alternative != nil {
		return Eval(ifStmt.Alternative)
	}

	return nil
}

func nativeBoolToBooleanObject(nodeVal bool) *object.Boolean {
	if nodeVal {
		return TRUE
	}

	return FALSE
}

// retEvalErrorObj is an error handling helper. Since evaluation functions
// expect some kind of object to be returned and we don't have nil, we
// just return EvalError object. The caller can decide what to do with it.
func retEvalErrorObj(message string) *object.EvalError {
	return &object.EvalError{Message: message}
}
