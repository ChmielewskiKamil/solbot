// object or "value" system used to evaluate the AST nodes.
package object

import (
	"fmt"
	"math/big"
)

type ObjectType string

const (
	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value big.Int
}

func (i *Integer) Inspect() string  { return i.Value.String() }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
