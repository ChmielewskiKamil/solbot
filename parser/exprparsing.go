package parser

import "solbot/ast"

// The exprparsing.go file contains the logic required to parse expressions.
// The technique used is called Pratt Parsing or top-down operator precedence
// parsing.

const (
	// Solidity's precedence table:
	// Numbers from 1 to 15 are taken from the Solidity's precedence table.
	_ int = iota
	// 15. Comma operator: ,
	LOWEST
	// 14. Ternary operator: ?
	// Assignment operators: =, |=, ^=, &=, <<=, >>=, +=, -=, *=, /=, %=
	TERNARY
	// 13. Logical OR operator: ||
	LOGICAL_OR
	// 12. Logical AND operator: &&
	LOGICAL_AND
	// 11. Equality operators: ==, !=
	EQUALITY
	// 10. Inequality operators: <, >, <=, >=
	INEQUALITY
	// 9. Bitwise OR operator: |
	BITWISE_OR
	// 8. Bitwise XOR operator: ^
	BITWISE_XOR
	// 7. Bitwise AND operator: &
	BITWISE_AND
	// 6. Shift operators: <<, >>
	BITWISE_SHIFT
	// 5. Addition and subtraction operators: +, -
	SUM
	// 4. Multiplication, division, and modulo operators: *, /, %
	PRODUCT
	// 3. Exponentiation operator: **
	EXPONENT
	// 2. Prefix incremend and decrement operators: ++, --
	// Unary minus: -
	// Unary operations: delete
	// Logical NOT: !
	// Bitwise NOT: ~
	PREFIX
	// 1. Postfix increment, decrement: ++, --
	// New expression: new <type name>
	// Array subscripting: <array>[<index>]
	// Member access: <object>.<member>
	// Function-like call: <function>(<args...>)
	// Parentheses: (<statement>)
	HIGHEST
)

// Inside the Parser struct we have two maps, prefixParseFns and infixParseFns.
// For each token type we can have two functions, one for parsing the prefix
// operators and one for parsing the infix operators.
type (
	prefixParseFn func() ast.Expression
	// infixParseFn accepts an argument that is the "left" side of the
	// infix operator that is being parsed.
	infixParseFn func(ast.Expression) ast.Expression
)

func (p *Parser) parseExpression(precedence int) ast.Expression {
	if p.trace {
		defer un(trace("parseExpression"))
	}

	prefix := p.prefixParseFns[p.currTkn.Type]
	if prefix == nil {
		return nil
	}

	leftExp := prefix()

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	if p.trace {
		defer un(trace("parseIdentifier"))
	}

	ident := &ast.Identifier{Pos: p.currTkn.Pos, Name: p.currTkn.Literal}
	return ident
}
