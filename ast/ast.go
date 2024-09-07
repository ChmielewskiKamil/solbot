package ast

import (
	"bytes"
	"math/big"
	"solbot/token"
)

// All nodes in the AST must implement the Node interface.
type Node interface {
	Start() token.Pos // First character of the node
	End() token.Pos   // First character immediately after the node
	String() string   // String representation of the node; helps debugging
}

// All expression nodes in the AST must implement the Expression interface.
type Expression interface {
	Node
	expressionNode()
}

// All statement nodes in the AST must implement the Statement interface.
type Statement interface {
	Node
	statementNode()
}

// All declaration nodes in the AST must implement the Declaration interface.
type Declaration interface {
	Node
	declarationNode()
}

// All type nodes in the AST must implement the Type interface.
type Type interface {
	Node
	typeNode()
}

/*~*~*~*~*~*~*~*~*~*~*~*~*~ Comments ~*~*~*~*~*~*~*~*~*~*~*~*~*~*/

type Comment struct {
	Slash token.Pos // Position of the leading '/'
	Text  string
}

func (c *Comment) Start() token.Pos { return c.Slash }
func (c *Comment) End() token.Pos {
	return token.Pos(int(c.Slash) + len(c.Text))
}

/*~*~*~*~*~*~*~*~*~*~ Expressions *~*~*~*~*~*~*~*~*~*~*/

type (
	Identifier struct {
		Pos   token.Pos // identifier position
		Value string    // identifier name
	}

	NumberLiteral struct {
		Pos   token.Pos   // position of the value
		Kind  token.Token // contains the token kind and literal string
		Value big.Int     // value of the literal; decimal or hex
	}

	BooleanLiteral struct {
		Pos   token.Pos // position of the value
		Value bool
	}

	PrefixExpression struct {
		Pos      token.Pos   // position of the operator
		Operator token.Token // operator token
		Right    Expression  // right operand
	}

	InfixExpression struct {
		Pos      token.Pos   // position of the left operand
		Left     Expression  // left operand
		Operator token.Token // operator token
		Right    Expression  // right operand
	}

	CallExpression struct {
		Pos      token.Pos    // Position of the identifier being called
		Function Expression   // Function being called
		Args     []Expression // Comma-separated list of arguments
	}
)

// Start() and End() implementations for Expression type Nodes

func (x *Identifier) Start() token.Pos { return x.Pos }
func (x *Identifier) End() token.Pos {
	return token.Pos(int(x.Pos) + len(x.Value))
}
func (x *NumberLiteral) Start() token.Pos { return x.Pos }
func (x *NumberLiteral) End() token.Pos {
	return token.Pos(int(x.Pos) + len(x.Kind.Literal))
}
func (x *BooleanLiteral) Start() token.Pos { return x.Pos }
func (x *BooleanLiteral) End() token.Pos {
	if x.Value {
		return token.Pos(int(x.Pos) + 4) // length of "true"
	}
	return token.Pos(int(x.Pos) + 5) // length of "false"
}
func (x *PrefixExpression) Start() token.Pos { return x.Pos }
func (x *PrefixExpression) End() token.Pos {
	return x.Right.End()
}
func (x *InfixExpression) Start() token.Pos { return x.Pos }
func (x *InfixExpression) End() token.Pos {
	return x.Right.End()
}
func (x *CallExpression) Start() token.Pos { return x.Pos }
func (x *CallExpression) End() token.Pos {
	if len(x.Args) > 0 {
		return x.Args[len(x.Args)-1].End() + 1 // @TODO: Shouldnt +2?
	}
	return x.Pos + 2 // length of "()"
}

// expressionNode() implementations to ensure that only expressions can be
// assigned to an Expression. This is useful if by mistake we try to use
// a Statement in a place where an Expression should be used instead.

func (*Identifier) expressionNode()       {}
func (*NumberLiteral) expressionNode()    {}
func (*BooleanLiteral) expressionNode()   {}
func (*PrefixExpression) expressionNode() {}
func (*InfixExpression) expressionNode()  {}
func (*CallExpression) expressionNode()   {}

// String() implementations for Expressions

func (x *Identifier) String() string    { return x.Value }
func (x *NumberLiteral) String() string { return x.Kind.Literal }
func (x *BooleanLiteral) String() string {
	if x.Value {
		return "true"
	}
	return "false"
}
func (x *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(x.Operator.Literal)
	out.WriteString(x.Right.String())
	out.WriteString(")")
	return out.String()
}
func (x *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(x.Left.String())
	out.WriteString(" ")
	out.WriteString(x.Operator.Literal)
	out.WriteString(" ")
	out.WriteString(x.Right.String())
	out.WriteString(")")
	return out.String()
}
func (x *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(x.Function.String())
	out.WriteString("(")
	for i, arg := range x.Args {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.String())
	}
	out.WriteString(")")
	return out.String()
}

/*~*~*~*~*~*~*~*~*~*~*~*~*~* Types ~*~*~*~*~*~*~*~*~*~*~*~*~*~*~*/
// Type nodes are constrains on expressions. They define the kinds of values
// that expressions can have. For example, ElementaryType constrains expressions
// to have values of a particular type e.g. "address", "uint256", "bool". In
// Solidity we have Five main types: elementary, function, user-defined, mapping
// and array types.

// In Solidity grammar called "ElementaryTypeName".
// One of: address, address payable, bool, string, uint, int, bytes,
// fixed, fixed-bytes or ufixed. NOT a Contract, Function, mapping (these are
// the four other types)
type ElementaryType struct {
	Pos  token.Pos   // position of the type keyword e.g. `a` in "address"
	Kind token.Token // type of the literal e.g. token.ADDRESS, token.UINT_256, token.BOOL
	// nil when used as declaration e.g. `address a;`
	// BUT also used in expressions e.g. `return uint256(a + b)`. Then it
	// contains the expression a + b
	Value Expression
}

// WARNING ElementaryType implements both Type and Expression interfaces.
// It can be used as a type e.g. in variable declaration `uint256 x;` OR
// as an expression in return statement `return uint256(a + b);`
func (x *ElementaryType) expressionNode() {}

// FunctionType represents a Solidity's function type. NOT TO BE CONFUSED WITH
// FUNCTION DECLARATION. FunctionType is a weird thing that no one uses (lol) e.g.
// ```solidity
//
//	struct Request {
//	    bytes data;
//	    function(uint) external callback; // <-- function type
//	}
//
//	// OR like this:
//	function query(bytes memory data, function(uint) external callback) public {
//	     requests.push(Request(data, callback));
//	     emit NewRequest(requests.length - 1);
//	}
//
// ```
type FunctionType struct {
	Pos        token.Pos  // position of the "function" keyword
	Params     *ParamList // input parameters; or nil
	Results    *ParamList // output parameters; or nil
	Mutability Mutability // mutability specifier e.g. pure, view, payable
	Visibility Visibility // visibility specifier e.g. public, private, internal, external
}

// @TODO: Implement user-defined type
// @TODO: Implement mapping type
// @TODO: Implement array types

// Param is not a type and not an expression, but we place it here since it is
// closely related to types.
type Param struct {
	Name         *Identifier  // param name e.g. "x" or "recipient"
	Type         Type         // e.g. ElementaryType, FunctionType etc.
	DataLocation DataLocation // data location e.g. storage, memory, calldata or nil
}

// ParamList is a list of parameters enclosed in parentheses. Similar to Param
// it is not a type or an expression, but we place it here since it is closely
// related to types.
type ParamList struct {
	Opening token.Pos // position of the opening parenthesis if any
	List    []*Param  // list of fields; or nil
	Closing token.Pos // position of the closing parenthesis if any
}

// @TODO: Implement Start(), End() and String() for FunctionType,
// Param and ParamList

// Start() and End() implementations for Expression type Nodes

func (x *ElementaryType) Start() token.Pos { return x.Pos }
func (x *ElementaryType) End() token.Pos   { return token.Pos(int(x.Pos) + int(x.End())) }

// expressionNode() implementations to ensure that only expressions and types
// can be assigned to an Expression. This is useful if by mistake we try to use
// a Statement in a place where an Expression should be used instead.

func (*ElementaryType) typeNode() {}

// String() implementations for Types
func (x *ElementaryType) String() string {
	var out bytes.Buffer
	out.WriteString(x.Kind.Literal)
	out.WriteString(" ")
	out.WriteString(x.Value.String())
	return out.String()
}

/*~*~*~*~*~*~*~*~*~*~*~*~* Statements *~*~*~*~*~*~*~*~*~*~*~*~*~*/

type (
	// In Solidity statements appear in blocks, which are enclosed in curly braces.
	// Block: { <<statement>> (and/or) <<unchecked-block>> }
	// For example: Constructor, Function, Modifier etc. delcarations have a body, which
	// is a block. Similarly try-catch, if-else, for, while statements have a block as well.
	BlockStatement struct {
		LeftBrace  token.Pos   // position of the left curly brace
		Statements []Statement // statements in the block
		RightBrace token.Pos   // position of the right curly brace
	}

	// UncheckedBlockStatement is a block that is declared as "unchecked".
	UncheckedBlockStatement struct {
		LeftBrace  token.Pos   // position of the left curly brace
		Statements []Statement // statements in the block
		RightBrace token.Pos   // position of the right curly brace
	}

	// VariableDeclarationStatement represents a declaration of a variable inside
	// a function. It is of a form: "type <<variable-name>> = <<expression>>;",
	// where the expression is optional.
	VariableDeclarationStatement struct {
		Type         Type         // e.g. elementary, function, user-defined type etc.
		Name         *Identifier  // variable name
		DataLocation DataLocation // data location e.g. storage, memory, calldata or nil
		Value        Expression   // initial value or nil; optional (NOT FOR TUPLES)
	}

	// @TODO Implement VariableDeclarationTupleStatement

	// Return statement is in a form of "return <<expression>>;", where
	// the expression is optional. In languages like Go, the return statement can
	// return an array of Expressions e.g., "return x, y, z". In Solidity, however,
	// if you want to return multiple values, you return a tuple-expression e.g.,
	// "return (x, y, z);".
	ReturnStatement struct {
		Pos    token.Pos  // position of the "return" keyword
		Result Expression // result expressions or nil
	}

	// ExpressionStatement is a statement that consists of a single expression.
	// It is of a form: <<expression>>;
	// In Solidity it is legal to write code like this: `x + 10;` or `x = 10;` or
	// `foo();`. In Go, unused expressions are not allowed e.g. `x + 10` will give
	// you an error.
	ExpressionStatement struct {
		Pos        token.Pos  // position of the first character of the expression
		Expression Expression // expression to be evaluated
	}

	IfStatement struct {
		Pos         token.Pos  // position of the "if" keyword
		Condition   Expression // condition to be evaluated
		Consequence Statement  // consequence happens if the condition is true; or nil
		Alternative Statement  // alternative happens if the condition is false; or nil
	}
)

// Start() and End() implementations for Statement type Nodes

func (s *BlockStatement) Start() token.Pos               { return s.LeftBrace }
func (s *BlockStatement) End() token.Pos                 { return s.RightBrace + 1 }
func (s *UncheckedBlockStatement) Start() token.Pos      { return s.LeftBrace }
func (s *UncheckedBlockStatement) End() token.Pos        { return s.RightBrace + 1 }
func (s *VariableDeclarationStatement) Start() token.Pos { return s.Type.Start() }
func (s *VariableDeclarationStatement) End() token.Pos   { return s.Value.End() }
func (s *ReturnStatement) Start() token.Pos              { return s.Pos }
func (s *ReturnStatement) End() token.Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Pos + 6 // length of "return"
}
func (s *ExpressionStatement) Start() token.Pos { return s.Pos }
func (s *ExpressionStatement) End() token.Pos   { return s.Expression.End() }
func (s *IfStatement) Start() token.Pos         { return s.Pos }
func (s *IfStatement) End() token.Pos {
	endPos := s.Pos + 2 // The length of "if".

	if s.Consequence != nil {
		endPos = s.Consequence.End()
	}

	if s.Alternative != nil {
		endPos = s.Alternative.End()
	}

	return endPos
}

// statementNode() ensures that only statement nodes can be assigned to a Statement.
func (*BlockStatement) statementNode()               {}
func (*UncheckedBlockStatement) statementNode()      {}
func (*VariableDeclarationStatement) statementNode() {}
func (*ReturnStatement) statementNode()              {}
func (*ExpressionStatement) statementNode()          {}
func (*IfStatement) statementNode()                  {}

// String() implementations for Statements

func (s *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	for _, stmt := range s.Statements {
		out.WriteString(stmt.String())
	}
	out.WriteString(" }")

	return out.String()
}

func (s *UncheckedBlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	for _, stmt := range s.Statements {
		out.WriteString(stmt.String())
	}
	out.WriteString(" }")

	return out.String()
}

func (s *VariableDeclarationStatement) String() string {
	var out bytes.Buffer
	out.WriteString(s.Type.String())
	out.WriteString(" ")
	out.WriteString(s.Name.String())
	if s.Value != nil {
		out.WriteString(" = ")
		out.WriteString(s.Value.String())
	}
	out.WriteString(";")

	return out.String()
}

func (s *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("return ")
	if s.Result != nil {
		out.WriteString(s.Result.String())
	}
	out.WriteString(";")

	return out.String()
}

func (s *ExpressionStatement) String() string {
	if s.Expression != nil {
		return s.Expression.String()
	}
	return ""
}

func (s *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(s.Condition.String())
	out.WriteString(" ")
	out.WriteString(s.Consequence.String())

	if s.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(s.Alternative.String())
	}

	return out.String()
}

/*~*~*~*~*~*~*~*~*~*~*~*~ Declarations ~*~*~*~*~*~*~*~*~*~*~*~*~*/

// @TODO: Add Contract declaration
// @TODO: Add Interface declaration
// @TODO: Add Library declaration
// @TODO: Add Struct declaration
// @TODO: Add Enum declaration
// @TODO: Add Event declaration
// @TODO: Add Error declaration
// @TODO: Add Using For Directive declaration
// @TODO: Add User Defined Value Type declaration

// Pragma and import directives could go into the File struct, since
// they are connected with a particular file.
// @TODO?: Add Pragma Directive declaration
// @TODO?: Add Import Directive declaration

// @TODO: Add modifier invocations *CallExpression
// @TODO: Add override specifier
// @TODO: Add documentation comments
type FunctionDeclaration struct {
	Pos        token.Pos       // position of the "function" keyword
	Name       *Identifier     // function name
	Params     *ParamList      // input parameters; or nil
	Results    *ParamList      // output parameters; or nil
	Mutability Mutability      // mutability specifier e.g. pure, view, payable
	Visibility Visibility      // visibility specifier e.g. public, private, internal, external
	Virtual    bool            // whether a function is marked as virtual
	Body       *BlockStatement // function body inside curly braces
}

// @TODO: State variables can have override specifier as well.
// StateVariableDeclaration represents a state variable declared inside a contract.
type StateVariableDeclaration struct {
	Name       *Identifier // variable name
	Type       Type        // e.g. ElementaryType
	Value      Expression  // initial value or nil
	Visibility Visibility  // visibility specifier: public, private, internal
	Mutability Mutability  // mutability specifier: constant, immutable, transient
}

// Start() and End() implementations for Declaration type Nodes

func (d *StateVariableDeclaration) Start() token.Pos { return d.Type.Start() }
func (d *StateVariableDeclaration) End() token.Pos   { return d.Value.End() }
func (d *FunctionDeclaration) Start() token.Pos      { return d.Name.Start() }
func (d *FunctionDeclaration) End() token.Pos        { return d.Body.End() }

// declarationNode() implementations to ensure that only declaration nodes can
// be assigned to a Declaration.

func (*StateVariableDeclaration) declarationNode() {}
func (*FunctionDeclaration) declarationNode()      {}

// String() implementations for Declarations
func (d *StateVariableDeclaration) String() string {
	// @TODO: If visibility and mutability is not set, they will give empty
	// spaces but who cares
	var out bytes.Buffer
	out.WriteString(d.Type.String())
	out.WriteString(" ")
	out.WriteString(d.Visibility.String())
	out.WriteString(" ")
	out.WriteString(d.Mutability.String())
	out.WriteString(" ")
	out.WriteString(d.Name.String())

	if d.Value != nil {
		out.WriteString(" = ")
		out.WriteString(d.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// @TODO: Implement String() for FunctionDeclaration
func (d *FunctionDeclaration) String() string { return "TO BE IMPLEMENTED" }

/*~*~*~*~*~*~*~*~*~*~*~*~*~* Files ~*~*~*~*~*~*~*~*~*~*~*~*~*~*~*/

// In Solidity grammar it's called "SourceUnit" and represents the entire source
// file.
type File struct {
	Name         string
	Declarations []Declaration
}

func (f *File) Start() token.Pos {
	if len(f.Declarations) > 0 {
		return f.Declarations[0].Start()
	}
	return 0
}

// @TODO: What if there is a trailing comment at the end?
func (f *File) End() token.Pos {
	if len(f.Declarations) > 0 {
		return f.Declarations[len(f.Declarations)-1].End()
	}
	return 0
}

func (f *File) String() string {
	var out bytes.Buffer
	for _, decl := range f.Declarations {
		out.WriteString(decl.String())
	}
	return out.String()
}

/*~*~*~*~*~*~ Visibility, Mutability, Data Location *~*~*~*~*~*~*/

// Visibility specifier for functions and function types. For convenience,
// this is also used for state variables. However, state vars can't be external.
type Visibility int

const (
	_ Visibility = iota
	Internal
	External
	Private
	Public
)

func (v Visibility) String() string {
	switch v {
	case Internal:
		return "internal"
	case External:
		return "external"
	case Private:
		return "private"
	case Public:
		return "public"
	default:
		return ""
	}
}

// State mutability specifier for functions. The default mutability of non-payable
// is assumed, if no mutability is specified. For convenience, constant,
// immutable and transient mutability specifiers are added for variables.
type Mutability int

const (
	_ Mutability = iota
	Pure
	View
	Payable
	Constant
	Immutable
	Transient
)

func (m Mutability) String() string {
	switch m {
	case Pure:
		return "pure"
	case View:
		return "view"
	case Payable:
		return "payable"
	case Constant:
		return "constant"
	case Immutable:
		return "immutable"
	case Transient:
		return "transient"
	default:
		return ""
	}
}

// Data location specifier for function parameter lists and variable declarations.
type DataLocation int

const (
	NO_DATA_LOCATION DataLocation = iota
	Storage
	Memory
	Calldata
)

func (d DataLocation) String() string {
	switch d {
	case Storage:
		return "storage"
	case Memory:
		return "memory"
	case Calldata:
		return "calldata"
	default:
		return ""
	}
}
