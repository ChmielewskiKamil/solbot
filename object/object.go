// object or "value" system used to evaluate the AST nodes.
package object

import (
	"fmt"
	"math/big"
)

type ObjectType string

const (
	EVAL_ERROR       = "EVAL_ERROR"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type (
	// Special object used to handle errors in object evaluation phase. It is
	// not connected to the Solidity language.
	EvalError struct {
		Message string
	}

	ReturnValue struct {
		Value Object
	}

	Integer struct {
		Value big.Int
	}

	Boolean struct {
		Value bool
	}
)

func (o *EvalError) Inspect() string {
	return fmt.Sprintf("Evaluation error: %s", o.Message)
}
func (o *EvalError) Type() ObjectType { return EVAL_ERROR }

func (o *ReturnValue) Inspect() string {
	return o.Value.Inspect()
}
func (o *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

func (o *Integer) Inspect() string  { return o.Value.String() }
func (o *Integer) Type() ObjectType { return INTEGER_OBJ }

func (o *Boolean) Inspect() string  { return fmt.Sprintf("%t", o.Value) }
func (o *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
