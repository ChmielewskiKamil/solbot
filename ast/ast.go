package ast

import "solbot/token"

// All nodes in the AST must implement the Node interface.
type Node interface {
	Start() token.Pos // First character of the node
	End() token.Pos   // First character immediately after the node
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

type Identifier struct {
	NamePos token.Pos // identifier position
	Name    string    // identifier name
}

// Start() and End() implementations for Expression type Nodes

func (x *Identifier) Start() token.Pos { return x.NamePos }
func (x *Identifier) End() token.Pos   { return token.Pos(int(x.NamePos) + len(x.Name)) }

// expressionNode() implementations to ensure that only expressions can be
// assigned to an Expression. This is useful if by mistake we try to use
// a Statement in a place where an Expression should be used instead.

func (*Identifier) expressionNode() {}

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
	Pos   token.Pos   // position of the type keyword e.g. `a` in "address"
	Kind  token.Token // type of the literal e.g. token.ADDRESS, token.UINT_256, token.BOOL
	Value string      // type literal value e.g. "address", "uint256", "bool" as a string
}

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
// @TODO: Data location is missing
type Param struct {
	Name *Identifier // param name e.g. "x" or "recipient"
	Type Type        // e.g. ElementaryType, FunctionType etc.
}

// ParamList is a list of parameters enclosed in parentheses. Similar to Param
// it is not a type or an expression, but we place it here since it is closely
// related to types.
type ParamList struct {
	Opening token.Pos // position of the opening parenthesis if any
	List    []*Param  // list of fields; or nil
	Closing token.Pos // position of the closing parenthesis if any
}

// Start() and End() implementations for Expression type Nodes

func (x *ElementaryType) Start() token.Pos { return x.Pos }
func (x *ElementaryType) End() token.Pos   { return token.Pos(int(x.Pos) + len(x.Value)) }

// expressionNode() implementations to ensure that only expressions and types
// can be assigned to an Expression. This is useful if by mistake we try to use
// a Statement in a place where an Expression should be used instead.

func (*ElementaryType) typeNode() {}

/*~*~*~*~*~*~*~*~*~*~*~*~* Statements *~*~*~*~*~*~*~*~*~*~*~*~*~*/

// In Solidity statements appear in blocks, which are enclosed in curly braces.
// Block: { <<statement>> (and/or) <<unchecked-block>> }
// For example: Constructor, Function, Modifier etc. delcarations have a body, which
// is a block. Similarly try-catch, if-else, for, while statements have a block as well.
type BlockStatement struct {
	LeftBrace  token.Pos   // position of the left curly brace
	Statements []Statement // statements in the block
	RightBrace token.Pos   // position of the right curly brace
}

// Return statement is in a form of "return <<expression>>;", where
// the expression is optional. In languages like Go, the return statement can
// return an array of Expressions e.g., "return x, y, z". In Solidity, however,
// if you want to return multiple values, you return a tuple-expression e.g.,
// "return (x, y, z);".
type ReturnStatement struct {
	Pos    token.Pos  // position of the "return" keyword
	Result Expression // result expressions or nil
}

// ExpressionStatement is a statement that consists of a single expression.
// It is of a form: <<expression>>;
// In Solidity it is legal to write code like this: `x + 10;` or `x = 10;` or
// `foo();`. In Go, unused expressions are not allowed e.g. `x + 10` will give
// you an error.
type ExpressionStatement struct {
	Pos        token.Pos  // position of the first character of the expression
	Expression Expression // expression to be evaluated
}

// Start() and End() implementations for Statement type Nodes

func (s *BlockStatement) Start() token.Pos  { return s.LeftBrace }
func (s *BlockStatement) End() token.Pos    { return s.RightBrace + 1 }
func (s *ReturnStatement) Start() token.Pos { return s.Pos }
func (s *ReturnStatement) End() token.Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Pos + 6 // length of "return"
}
func (s *ExpressionStatement) Start() token.Pos { return s.Pos }
func (s *ExpressionStatement) End() token.Pos   { return s.Expression.End() }

// statementNode() ensures that only statement nodes can be assigned to a Statement.
func (*BlockStatement) statementNode()      {}
func (*ReturnStatement) statementNode()     {}
func (*ExpressionStatement) statementNode() {}

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
// @TODO: Add documentation comments
type FunctionDeclaration struct {
	Name *Identifier     // function name
	Type *FunctionType   // function signature with input/output parameters, mutability, visibility
	Body *BlockStatement // function body inside curly braces
}

// @TODO: There is the new `transient` keyword
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

// Data location specifier for function parameter lists and variable declarations.
type DataLocation int

const (
	_ DataLocation = iota
	Storage
	Memory
	Calldata
)
