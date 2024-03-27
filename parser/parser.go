package parser

import (
	"solparsor/ast"
	"solparsor/lexer"
	"solparsor/token"
)

type Parser struct {
	l lexer.Lexer

	currTkn token.Token
	peekTkn token.Token
}

func (p *Parser) init(src string) {
	p.l = *lexer.Lex(src)
	p.nextToken()
	p.nextToken()
}

func (p *Parser) nextToken() {
	p.currTkn = p.peekTkn
	p.peekTkn = p.l.NextToken()
}

func (p *Parser) ParseFile() *ast.File {
	return nil
}
