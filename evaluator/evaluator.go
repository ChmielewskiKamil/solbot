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

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	default:
		return newError(fmt.Sprintf("Unhandled ast node: %T", node))

	// File

	case *ast.File:
		return evalDeclarations(node.Declarations, env)

	// Declarations

	case *ast.FunctionDeclaration:
		return evalFunctionDeclaration(node, env)

	// Statements

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.IfStatement:
		return evalIfStatement(node, env)
	case *ast.VariableDeclarationStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return env.Set(node.Name.Value, val)

	// Expressions

	case *ast.NumberLiteral:
		// TODO: This should probably distinguish between hex and decimal.
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.ReturnStatement:
		result := Eval(node.Result, env)
		if isError(result) {
			return result
		}
		return &object.ReturnValue{Value: result}
	case *ast.Identifier:
		return evalIdentifier(node, env)
	}
}

func evalFunctionDeclaration(fn *ast.FunctionDeclaration, env *object.Environment) object.Object {
	function := &object.Function{
		Name: fn.Name,
		Body: fn.Body,
		Env:  env,
	}

	return env.Set(function.Name.Value, function)
}

func evalDeclarations(decls []ast.Declaration, env *object.Environment) object.Object {
	var result object.Object

	for _, decl := range decls {
		result = Eval(decl, env)
		if isError(result) {
			return result
		}
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)
		if result != nil {
			// If we encounter a return statement (return value object) just bubble
			// it up and let the outer block handle it. Same for error.
			if result.Type() == object.EVAL_ERROR_OBJ || result.Type() == object.RETURN_VALUE_OBJ {
				return result
			}
		}
	}

	return result
}

func evalPrefixExpression(operator token.Token, right object.Object) object.Object {
	switch operator.Type {
	default:
		return newError("Unknown prefix operator '%s%s'.",
			operator.Literal, right.Type())
	case token.NOT:
		return evalNotPrefixOperatorExpression(right)
	case token.SUB:
		return evalSubPrefixOperatorExpression(right)
	}
}

func evalNotPrefixOperatorExpression(right object.Object) object.Object {
	switch right {
	default:
		return newError(
			"The '!' prefix operator can only be used with booleans. Got: %s instead.", right.Type(),
		)
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	}
}

func evalSubPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("The '-' prefix operator can only be used with integers. Got: %s instead.",
			right.Type())
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
		return newError("Incorrect object types for infix expression: %s %s %s.",
			left.Type(), operator.Literal, right.Type())
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
		return newError(
			"Incorrect operator: '%s' in integer infix expression.",
			operator.Literal)
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

func evalIfStatement(ifStmt *ast.IfStatement, env *object.Environment) object.Object {
	evaluated := Eval(ifStmt.Condition, env)
	cond, ok := evaluated.(*object.Boolean)
	if !ok {
		return newError(
			"The condition has to be an *object.Boolean, got: %T instead.",
			evaluated)
	}

	if cond.Value == true {
		return Eval(ifStmt.Consequence, env)
	} else if ifStmt.Alternative != nil {
		return Eval(ifStmt.Alternative, env)
	}

	return nil
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("Identifier not found: %s", node.Value)
	}

	return val
}

func nativeBoolToBooleanObject(nodeVal bool) *object.Boolean {
	if nodeVal {
		return TRUE
	}

	return FALSE
}

// newError is an error handling helper. Since evaluation functions
// expect some kind of object to be returned and we don't have nil, we
// just return EvalError object. The caller can decide what to do with it.
func newError(format string, a ...interface{}) *object.EvalError {
	return &object.EvalError{Message: fmt.Sprintf(format, a...)}
}

// isError is used to stop evaluation early when calling Eval recursively.
// For example when evaluating left + right, there is no need to continue Eval
// when we know that left returned an error.
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.EVAL_ERROR_OBJ
	}
	return false
}
