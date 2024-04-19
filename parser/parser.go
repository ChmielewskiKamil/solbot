package parser

import (
	"fmt"
	"solparsor/ast"
	"solparsor/lexer"
	"solparsor/token"
)

type Parser struct {
	l      lexer.Lexer
	errors ErrorList

	currTkn token.Token
	peekTkn token.Token
}

func (p *Parser) init(src string) {
	p.l = *lexer.Lex(src)
	p.errors = ErrorList{}

	// Read two tokens, so currTkn and peekTkn are both set
	p.nextToken()
	p.nextToken()
}

func (p *Parser) nextToken() {
	p.currTkn = p.peekTkn
	p.peekTkn = p.l.NextToken()
}

func (p *Parser) ParseFile() *ast.File {
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
	switch p.currTkn.Type {
	case token.ADDRESS, token.UINT_256, token.BOOL:
		return p.parseVariableDeclaration()
	default:
		return nil
	}
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	decl := &ast.VariableDeclaration{}

	// We are sitting on the variable type e.g. address or uint256
	decl.Type = &ast.ElementaryType{
		ValuePos: p.currTkn.Pos,
		Kind:     p.currTkn,
		Value:    p.currTkn.Literal,
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
