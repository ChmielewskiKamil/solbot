// object or "value" system used to evaluate the AST nodes.
package object

import (
	"bytes"
	"fmt"
	"math/big"
	"solbot/ast"
	"strings"
)

type ObjectType string

const (
	EVAL_ERROR_OBJ   = "EVAL_ERROR"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	FUNCTION_OBJ     = "FUNCTION"
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

	Function struct {
		Name       *ast.Identifier
		Params     *ast.ParamList
		Results    *ast.ParamList
		Mutability ast.Mutability
		Visibility ast.Visibility
		Virtual    bool
		Body       *ast.BlockStatement
		Env        *Environment // TODO: Do I need the env if Solidity does not have closures?
	}
)

func (o *EvalError) Inspect() string {
	return fmt.Sprintf("Evaluation error: %s", o.Message)
}
func (o *EvalError) Type() ObjectType { return EVAL_ERROR_OBJ }

func (o *ReturnValue) Inspect() string {
	return o.Value.Inspect()
}
func (o *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

func (o *Integer) Inspect() string  { return o.Value.String() }
func (o *Integer) Type() ObjectType { return INTEGER_OBJ }

func (o *Boolean) Inspect() string  { return fmt.Sprintf("%t", o.Value) }
func (o *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

func (o *Function) Inspect() string {
	var out bytes.Buffer

	// non-nil slice; it is initialized so we can append elements right away
	// if there are no params this will just be empty string
	params := []string{}
	for _, p := range o.Params.List {
		var str string
		if p.DataLocation != ast.NO_DATA_LOCATION {
			str = p.Type.String() + " " + p.DataLocation.String() + " " + p.Name.String()
		} else {
			str = p.Type.String() + " " + p.Name.String()
		}
		params = append(params, str)
	}

	results := []string{}
	for _, r := range o.Results.List {
		var str string
		if r.DataLocation != ast.NO_DATA_LOCATION {
			str = r.Type.String() + " " + r.DataLocation.String() + " " + r.Name.String()
		} else {
			str = r.Type.String() + " " + r.Name.String()
		}
		results = append(results, str)
	}

	out.WriteString("function")
	out.WriteString(o.Name.String())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")\n")
	// TODO: Add mutability, visibility, virtual and return params
	out.WriteString("{\n")
	out.WriteString(o.Body.String())
	out.WriteString("}")

	return out.String()
}

func (o *Function) Type() ObjectType { return FUNCTION_OBJ }
