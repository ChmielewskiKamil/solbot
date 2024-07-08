package parser

import (
	"fmt"
	"solbot/ast"
	"solbot/lexer"
	"solbot/token"
)

type Parser struct {
	file   *token.File
	l      lexer.Lexer
	errors ErrorList

	// Tracing
	trace bool

	currTkn token.Token
	peekTkn token.Token
}

func (p *Parser) Init(file *token.File) {
	p.l = *lexer.Lex(file)
	p.errors = ErrorList{}
	p.file = file
	p.trace = false

	// Read two tokens, so currTkn and peekTkn are both set
	p.nextToken()
	p.nextToken()
}

func (p *Parser) ToggleTracing() {
	p.trace = !p.trace
}

func (p *Parser) nextToken() {
	p.currTkn = p.peekTkn
	p.peekTkn = p.l.NextToken()
}

func (p *Parser) ParseFile() *ast.File {
	if p.trace {
		defer un(trace("ParseFile"))
	}

	file := &ast.File{}
	file.Declarations = []ast.Declaration{}

	for p.currTkn.Type != token.EOF {
		decl := p.parseDeclaration()
		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
		p.nextToken()
	}

	return file
}

func (p *Parser) parseDeclaration() ast.Declaration {
	if p.trace {
		defer un(trace("parseDeclaration"))
	}
	switch tkType := p.currTkn.Type; {
	case token.IsElementaryType(tkType):
		return p.parseStateVariableDeclaration()
	case tkType == token.FUNCTION:
		return p.parseFunctionDeclaration()
	default:
		return nil
	}
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	if p.trace {
		defer un(trace("parseFunctionDeclaration"))
	}
	decl := &ast.FunctionDeclaration{}

	// 1. Function keyword
	fnType := &ast.FunctionType{}
	fnType.Func = p.currTkn.Pos

	// 2. Function identifier
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	decl.Name = &ast.Identifier{
		NamePos: p.currTkn.Pos,
		Name:    p.currTkn.Literal,
	}

	// 3. ( Param List )
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	params := &ast.ParamList{}
	params.Opening = p.currTkn.Pos

	for !p.currTknIs(token.RPAREN) {
		p.nextToken()
	}

	params.Closing = p.currTkn.Pos

	// 4. Visibility, State Mutability, Modifier Invocation, Override, Virtual

	// 5. Returns ( Param List )

	// 6. Body block
	for !p.currTknIs(token.LBRACE) {
		p.nextToken()
	}

	fnBody := p.parseBlockStatement()

	// 7. Semicolon

	fnType.Params = params
	decl.Body = fnBody
	decl.Type = fnType
	return decl
}

func (p *Parser) parseStateVariableDeclaration() *ast.StateVariableDeclaration {
	if p.trace {
		defer un(trace("parseStateVariableDeclaration"))
	}
	decl := &ast.StateVariableDeclaration{}

	// Set default values so that we don't have nil pointer dereferences
	decl.Visibility = ast.Internal // Solidity default visibility

	// decl.Mutability does not need to be set since all variables are mutable
	// by default, and here we can only set Constant, Immutable, or Transient.

	// We are sitting on the variable type e.g. address or uint256
	decl.Type = &ast.ElementaryType{
		ValuePos: p.currTkn.Pos,
		Kind:     p.currTkn,
		Value:    p.currTkn.Literal,
	}

	p.nextToken()
	// We might be sitting on the variable name OR the visibility specifier OR the mutability specifier

	// Visibility and mutability specifiers are flexible. Both are valid:
	// uint256 public constant x = 10;
	// uint256 constant public x = 10;

	for {
		switch tkType := p.currTkn.Type; {
		default:
			goto skip_expression
		case token.IsVarVisibility(tkType):
			switch tkType {
			case token.PUBLIC:
				decl.Visibility = ast.Public
			case token.PRIVATE:
				decl.Visibility = ast.Private
			case token.INTERNAL:
				decl.Visibility = ast.Internal
			}
		case token.IsVarMutability(tkType):
			switch tkType {
			case token.CONSTANT:
				decl.Mutability = ast.Constant
			case token.IMMUTABLE:
				decl.Mutability = ast.Immutable
			case token.TRANSIENT:
				decl.Mutability = ast.Transient
			}
		case tkType == token.OVERRIDE:
			// @TODO: Handle override. It requires changes in the AST.
			p.nextToken()
			continue
		case tkType == token.IDENTIFIER:
			decl.Name = &ast.Identifier{
				NamePos: p.currTkn.Pos,
				Name:    p.currTkn.Literal,
			}
		case tkType == token.ASSIGN:
			// @TODO: The next token is an expression, which we want to skip
			// for now.
			p.nextToken()
			goto skip_expression
		}
		p.nextToken()
	}

skip_expression:
	// @TODO: We skip the Value for now since it is an expression.
	// The variable declaration ends with a semicolon.
	for !p.currTknIs(token.SEMICOLON) {
		p.nextToken()
	}

	return decl
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	if p.trace {
		defer un(trace("parseBlockStatement"))
	}
	blockStmt := &ast.BlockStatement{}
	blockStmt.LeftBrace = p.currTkn.Pos

	for !p.currTknIs(token.RBRACE) {
		p.nextToken()
	}

	blockStmt.RightBrace = p.currTkn.Pos

	return blockStmt
}

// expectPeek checks if the next token is of the expected type.
// If it is it advances the tokens.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTkn.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be: %s, got: %s instead (at offset: %d)",
		t.String(), p.peekTkn.Type.String(), p.peekTkn.Pos)
	p.errors.Add(p.peekTkn.Pos, msg)
}

// currTknIs checks if the current token is of the expected type.
func (p *Parser) currTknIs(t token.TokenType) bool {
	return p.currTkn.Type == t
}
