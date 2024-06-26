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
		return p.parseVariableDeclaration()
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

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	if p.trace {
		defer un(trace("parseVariableDeclaration"))
	}
	decl := &ast.VariableDeclaration{}

	// Set default values so that we don't have nil pointer dereferences
	decl.Constant = false

	// We are sitting on the variable type e.g. address or uint256
	decl.Type = &ast.ElementaryType{
		ValuePos: p.currTkn.Pos,
		Kind:     p.currTkn,
		Value:    p.currTkn.Literal,
	}

	// @TODO: We need to handle visibility
	if isVisibility(p.peekTkn.Type) {

	}

	if p.peekTkn.Type == token.CONSTANT {
		decl.Constant = true
		p.nextToken()
	}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	decl.Name = &ast.Identifier{
		NamePos: p.currTkn.Pos,
		Name:    p.currTkn.Literal,
	}

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
