package parser

import (
	"github.com/ChmielewskiKamil/solbot/ast"
	"github.com/ChmielewskiKamil/solbot/token"
	"math/big"
)

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

// Solidity's precedence table LOW (15) -> HIGH (1)
var precedences = map[token.TokenType]int{
	// 15.
	token.COMMA: LOWEST,
	// 14.
	token.CONDITIONAL:    TERNARY,
	token.ASSIGN:         TERNARY,
	token.ASSIGN_BIT_OR:  TERNARY,
	token.ASSIGN_BIT_XOR: TERNARY,
	token.ASSIGN_BIT_AND: TERNARY,
	token.ASSIGN_SHL:     TERNARY,
	token.ASSIGN_SAR:     TERNARY,
	token.ASSIGN_SHR:     TERNARY, // TODO: It is in language grammar, but not in the precedence cheatsheet, why?
	token.ASSIGN_ADD:     TERNARY,
	token.ASSIGN_SUB:     TERNARY,
	token.ASSIGN_MUL:     TERNARY,
	token.ASSIGN_DIV:     TERNARY,
	token.ASSIGN_MOD:     TERNARY,
	// 13.
	token.OR: LOGICAL_OR,
	// 12.
	token.AND: LOGICAL_AND,
	// 11.
	token.EQUAL:     EQUALITY,
	token.NOT_EQUAL: EQUALITY,
	// 10.
	token.LESS_THAN:             INEQUALITY,
	token.GREATER_THAN:          INEQUALITY,
	token.LESS_THAN_OR_EQUAL:    INEQUALITY,
	token.GREATER_THAN_OR_EQUAL: INEQUALITY,
	// 9.
	token.BIT_OR: BITWISE_OR,
	// 8.
	token.BIT_XOR: BITWISE_XOR,
	// 7.
	token.BIT_AND: BITWISE_AND,
	// 6.
	token.SAR: BITWISE_SHIFT,
	token.SHL: BITWISE_SHIFT,
	// 5.
	token.ADD: SUM,
	token.SUB: SUM,
	// 4.
	token.MUL: PRODUCT,
	token.DIV: PRODUCT,
	token.MOD: PRODUCT,
	// 3.
	token.EXP: EXPONENT,
	// 2.
	token.INC: PREFIX,
	token.DEC: PREFIX,
	// token.SUB: PREFIX, // TODO: UNARY MINUS is problematic
	token.DELETE:  PREFIX,
	token.NOT:     PREFIX,
	token.BIT_NOT: PREFIX,
	// 1.
	// TODO: The function calls, array subscripting, member access etc.
	// is harder to implement.
	token.LPAREN: HIGHEST,
}

func (p *parser) peekPrecedence() int {
	if p, ok := precedences[p.peekTkn.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *parser) currPrecedence() int {
	if p, ok := precedences[p.currTkn.Type]; ok {
		return p
	}
	return LOWEST
}

// Inside the parser struct we have two maps, prefixParseFns and infixParseFns.
// For each token type we can have two functions, one for parsing the prefix
// operators and one for parsing the infix operators.
type (
	prefixParseFn func() ast.Expression
	// infixParseFn accepts an argument that is the "left" side of the
	// infix operator that is being parsed.
	infixParseFn func(ast.Expression) ast.Expression
)

func (p *parser) parseExpression(precedence int) ast.Expression {
	if p.trace {
		defer un(trace("parseExpression"))
	}

	prefix := p.prefixParseFns[p.currTkn.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currTkn.Type)
		return nil
	}

	leftExp := prefix()

	for p.peekTkn.Type != token.SEMICOLON && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekTkn.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *parser) noPrefixParseFnError(t token.TokenType) {
	msg := "no prefix parse function for '" + t.String() + "' found"
	p.addError(p.currTkn.Pos, msg)
}

func (p *parser) parseIdentifier() ast.Expression {
	if p.trace {
		defer un(trace("parseIdentifier"))
	}

	ident := &ast.Identifier{Pos: p.currTkn.Pos, Value: p.currTkn.Literal}
	return ident
}

func (p *parser) parseNumberLiteral() ast.Expression {
	if p.trace {
		defer un(trace("parseNumberLiteral"))
	}

	numLit := &ast.NumberLiteral{
		Pos: p.currTkn.Pos,
		Kind: token.Token{
			Type: p.currTkn.Type, Literal: p.currTkn.Literal, Pos: p.currTkn.Pos,
		}}

	bigInt, ok := new(big.Int).SetString(p.currTkn.Literal, 0)
	if !ok {
		p.addError(p.currTkn.Pos, "could not parse number literal")
		return nil
	}

	numLit.Value = *bigInt

	return numLit
}

func (p *parser) parseBooleanLiteral() ast.Expression {
	if p.trace {
		defer un(trace("parseBooleanLiteral"))
	}

	bl := &ast.BooleanLiteral{
		Pos:   p.currTkn.Pos,
		Value: p.currTknIs(token.TRUE_LITERAL),
	}
	return bl
}

func (p *parser) parsePrefixExpression() ast.Expression {
	if p.trace {
		defer un(trace("parsePrefixExpression"))
	}

	pe := &ast.PrefixExpression{
		Pos: p.currTkn.Pos,
		Operator: token.Token{
			Type:    p.currTkn.Type,
			Literal: p.currTkn.Literal,
			Pos:     p.currTkn.Pos},
	}

	p.nextToken()

	pe.Right = p.parseExpression(PREFIX)

	return pe
}

func (p *parser) parseInfixExpression(left ast.Expression) ast.Expression {
	if p.trace {
		defer un(trace("parseInfixExpression"))
	}

	exp := &ast.InfixExpression{
		Pos:  left.Start(),
		Left: left,
		Operator: token.Token{
			Type:    p.currTkn.Type,
			Literal: p.currTkn.Literal,
			Pos:     p.currTkn.Pos,
		},
	}

	precedence := p.currPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)

	return exp
}

func (p *parser) parseGroupedExpression() ast.Expression {
	if p.trace {
		defer un(trace("parseGroupedExpression"))
	}

	p.nextToken()

	// Parse the thing inside parentheses.
	exp := p.parseExpression(LOWEST)

	// There should be a closing parenthesis.
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *parser) parseElementaryTypeExpression() ast.Expression {
	if p.trace {
		defer un(trace("parseElementaryType"))
	}

	et := &ast.ElementaryTypeExpression{
		Pos: p.currTkn.Pos,
		Kind: token.Token{
			Type:    p.currTkn.Type,
			Literal: p.currTkn.Literal,
			Pos:     p.currTkn.Pos,
		},
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	et.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Semi-colon is optional.
	if p.peekTkn.Type == token.SEMICOLON {
		p.nextToken()
	}

	return et
}

// parseCallExpression is called when LPAREN is encountered in an
// infix position. That's why we have access to the function ident
// and we can pass it as an argument to the function.
func (p *parser) parseCallExpression(fn ast.Expression) ast.Expression {
	if p.trace {
		defer un(trace("parseCallExpression"))
	}

	callExp := &ast.CallExpression{
		Pos:   fn.Start(), // TODO: fn is not passed as ref, will this work?
		Ident: fn,
	}

	callExp.Args = p.parseCallArguments()
	return callExp
}

func (p *parser) parseCallArguments() []ast.Expression {
	if p.trace {
		defer un(trace("parseCallArguments"))
	}

	args := []ast.Expression{}
	if p.peekTkn.Type == token.RPAREN {
		p.nextToken()
		return args
	}

	// Move to the first arg.
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTkn.Type == token.COMMA {
		p.nextToken() // Sitting on comma
		p.nextToken() // Move to the next arg

		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}
