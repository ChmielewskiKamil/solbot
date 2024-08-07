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

	// Pratt Parsing maps are used to parse expressions. They define the logic
	// on how to parse a specific token based on its position.
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) Init(file *token.File) {
	p.l = *lexer.Lex(file)
	p.errors = ErrorList{}
	p.file = file
	p.trace = false

	// Read two tokens, so currTkn and peekTkn are both set
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENTIFIER, p.parseIdentifier)
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
	decl.Pos = p.currTkn.Pos

	// 2. Function identifier
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	decl.Name = &ast.Identifier{
		Pos:  p.currTkn.Pos,
		Name: p.currTkn.Literal,
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

	decl.Params = params
	decl.Body = fnBody
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
		Pos:   p.currTkn.Pos,
		Kind:  p.currTkn,
		Value: p.currTkn.Literal,
	}

	p.nextToken()

	// We might be sitting on the variable name OR the visibility specifier OR the mutability specifier

	// Visibility and mutability specifiers are flexible. Both are valid:
	// uint256 public constant x = 10;
	// uint256 constant public x = 10;

	for {
		switch tkType := p.currTkn.Type; {
		default:
			p.nextToken()
		case tkType == token.IDENTIFIER:
			decl.Name = &ast.Identifier{
				Pos:  p.currTkn.Pos,
				Name: p.currTkn.Literal,
			}
			p.nextToken()
		case token.IsVarVisibility(tkType):
			switch tkType {
			case token.PUBLIC:
				decl.Visibility = ast.Public
			case token.PRIVATE:
				decl.Visibility = ast.Private
			case token.INTERNAL:
				decl.Visibility = ast.Internal
			}
			p.nextToken()
		case token.IsVarMutability(tkType):
			switch tkType {
			case token.CONSTANT:
				decl.Mutability = ast.Constant
			case token.IMMUTABLE:
				decl.Mutability = ast.Immutable
			case token.TRANSIENT:
				decl.Mutability = ast.Transient
			}
			p.nextToken()
		case tkType == token.OVERRIDE:
			// @TODO: Handle override. It requires changes in the AST.
			p.nextToken()
		case tkType == token.ASSIGN:
			// @TODO: The next token is an expression, which we want to skip
			// for now.
			for !p.currTknIs(token.SEMICOLON) {
				p.nextToken()
			}
		case tkType == token.SEMICOLON:
			return decl
		}
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch tkType := p.currTkn.Type; {
	default:
		return p.parseExpressionStatement()
	case token.IsElementaryType(tkType):
		// @TODO: Implement other types that variables can have.
		return p.parseVariableDeclarationStatement()
	case tkType == token.LBRACE:
		return p.parseBlockStatement()
	case tkType == token.RETURN:
		return p.parseReturnStatement()
	}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	if p.trace {
		defer un(trace("parseBlockStatement"))
	}
	blockStmt := &ast.BlockStatement{}
	blockStmt.LeftBrace = p.currTkn.Pos

	p.nextToken()

	for {
		switch tkType := p.currTkn.Type; {
		default:
			stmt := p.parseStatement()
			if stmt != nil {
				blockStmt.Statements = append(blockStmt.Statements, stmt)
			}
			// When we parse a statement e.g. if statement, at the end we will
			// land at the RBRACE ending the if statement. To ensure that we
			// handle the scope of the current block correctly, we advance
			// by one token. This way the only encountered RBRACE in this for
			// loop will be the end of the current block.
			p.nextToken()
		case tkType == token.UNCHECKED:
			// Move to the LBRACE.
			p.nextToken()
			stmt := p.parseUncheckedBlockStatement()
			if stmt != nil {
				blockStmt.Statements = append(blockStmt.Statements, stmt)
			}
			// Block parsing ends on RBRACE, so advance to the next token.
			p.nextToken()
		case tkType == token.RBRACE:
			// We have reached the end of the block.
			blockStmt.RightBrace = p.currTkn.Pos
			return blockStmt
		}
	}
}

// Almost the same as parseBlockStatement, but we don't allow nested unchecked
func (p *Parser) parseUncheckedBlockStatement() *ast.UncheckedBlockStatement {
	if p.trace {
		defer un(trace("parseUncheckedBlockStatement"))
	}
	blockStmt := &ast.UncheckedBlockStatement{}
	blockStmt.LeftBrace = p.currTkn.Pos

	p.nextToken()
	for {
		switch tkType := p.currTkn.Type; {
		default:
			stmt := p.parseStatement()
			blockStmt.Statements = append(blockStmt.Statements, stmt)
			// When we parse a statement e.g. if statement, at the end we will
			// land at the RBRACE ending the if statement. To ensure that we
			// handle the scope of the current block correctly, we advance
			// by one token. This way the only encountered RBRACE in this for
			// loop will be the end of the current block.
			p.nextToken()
		case tkType == token.UNCHECKED:
			p.errors.Add(p.currTkn.Pos, "Nested unchecked blocks are not allowed.")
			p.nextToken() // Consume the offending token and continue parsing
			return nil
		case tkType == token.RBRACE:
			// We have reached the end of the block.
			blockStmt.RightBrace = p.currTkn.Pos
			return blockStmt
		}
	}
}

func (p *Parser) parseVariableDeclarationStatement() *ast.VariableDeclarationStatement {
	if p.trace {
		defer un(trace("parseVariableDeclarationStatement"))
	}
	vdStmt := &ast.VariableDeclarationStatement{}
	vdStmt.DataLocation = ast.NO_DATA_LOCATION // assign default value

	vdStmt.Type = &ast.ElementaryType{
		Pos:   p.currTkn.Pos,
		Kind:  p.currTkn,
		Value: p.currTkn.Literal,
	}

	p.nextToken()

	for {
		switch tkType := p.currTkn.Type; {
		default:
			p.nextToken()
		case tkType == token.IDENTIFIER:
			vdStmt.Name = &ast.Identifier{
				Pos:  p.currTkn.Pos,
				Name: p.currTkn.Literal,
			}
			p.nextToken()
		case token.IsDataLocation(tkType):
			switch tkType {
			case token.STORAGE:
				vdStmt.DataLocation = ast.Storage
			case token.MEMORY:
				vdStmt.DataLocation = ast.Memory
			case token.CALLDATA:
				vdStmt.DataLocation = ast.Calldata
			}
			p.nextToken()
		case tkType == token.SEMICOLON:
			return vdStmt
		}
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	if p.trace {
		defer un(trace("parseReturnStatement()"))
	}

	retStmt := &ast.ReturnStatement{}
	retStmt.Pos = p.currTkn.Pos

	// @TODO: Parse expression
	for !p.currTknIs(token.SEMICOLON) {
		p.nextToken()
	}

	return retStmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	if p.trace {
		defer un(trace("parseExpressionStatement"))
	}
	exprStmt := &ast.ExpressionStatement{}
	exprStmt.Pos = p.currTkn.Pos

	exprStmt.Expression = p.parseExpression(LOWEST)

	// Semicolon is optional on purpose to make it easier to type stuff into
	// the REPL like: 1 + 2 without the need for the semicolon.
	if p.peekTkn.Type == token.SEMICOLON {
		p.nextToken()
	}

	return exprStmt
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

func (p *Parser) registerPrefix(t token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

func (p *Parser) registerInfix(t token.TokenType, fn infixParseFn) {
	p.infixParseFns[t] = fn
}
