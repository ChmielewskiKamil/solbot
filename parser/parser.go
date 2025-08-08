package parser

import (
	"fmt"
	"io"

	"github.com/ChmielewskiKamil/solbot/ast"
	"github.com/ChmielewskiKamil/solbot/lexer"
	"github.com/ChmielewskiKamil/solbot/token"
)

////////////////////////////////////////////////////
//////////////////// PUBLIC API ////////////////////
////////////////////////////////////////////////////

// ParseFile parses the Solidity source code from src and returns the corresponding AST.
// The filename parameter is used for context in error messages and position tracking.
// Optional configuration, such as tracing, can be provided via the opts parameter.
//
// If parsing fails with syntax errors, the function returns a non-nil *ast.File
// (which may be partially complete) and an error of type ErrorList.
func ParseFile(filename string, src io.Reader, opts ...Option) (*ast.File, error) {
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the file %s: %w", filename, err)
	}

	sourceFile, err := token.NewSourceFile(filename, string(content))
	if err != nil {
		return nil, fmt.Errorf("Failed to create new source file for %s: %w", filename, err)
	}

	p := newParser(sourceFile)

	for _, opt := range opts {
		opt(p)
	}

	file := p.parseFile()

	if len(p.errors) > 0 {
		return file, p.errors
	}

	return file, nil
}

type Option func(*parser)

func WithTracing() Option {
	return func(p *parser) {
		p.trace = true
	}
}

////////////////////////////////////////////////////
//////////////////// INTERNALS /////////////////////
////////////////////////////////////////////////////

type parser struct {
	file   *token.SourceFile
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

func newParser(file *token.SourceFile) *parser {
	p := &parser{
		file:   file,
		l:      *lexer.Lex(file),
		errors: ErrorList{},
	}

	p.trace = false

	// Read two tokens, so currTkn and peekTkn are both set
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(token.DECIMAL_NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.HEX_NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.TRUE_LITERAL, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE_LITERAL, p.parseBooleanLiteral)

	registerPrefixElementaryTypes(p)

	// Prefix Expressions
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.BIT_NOT, p.parsePrefixExpression)
	p.registerPrefix(token.INC, p.parsePrefixExpression)
	p.registerPrefix(token.DEC, p.parsePrefixExpression)
	p.registerPrefix(token.DELETE, p.parsePrefixExpression)
	p.registerPrefix(token.SUB, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	// Infix Expressions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.EXP, p.parseInfixExpression)
	p.registerInfix(token.MUL, p.parseInfixExpression)
	p.registerInfix(token.DIV, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.ADD, p.parseInfixExpression)
	p.registerInfix(token.SUB, p.parseInfixExpression)
	p.registerInfix(token.SAR, p.parseInfixExpression)
	p.registerInfix(token.SHL, p.parseInfixExpression)
	p.registerInfix(token.SHR, p.parseInfixExpression)
	p.registerInfix(token.BIT_AND, p.parseInfixExpression)
	p.registerInfix(token.BIT_XOR, p.parseInfixExpression)
	p.registerInfix(token.BIT_OR, p.parseInfixExpression)
	p.registerInfix(token.LESS_THAN, p.parseInfixExpression)
	p.registerInfix(token.GREATER_THAN, p.parseInfixExpression)
	p.registerInfix(token.LESS_THAN_OR_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.GREATER_THAN_OR_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.EQUAL, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.CONDITIONAL, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_BIT_OR, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_BIT_XOR, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_BIT_AND, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_SHL, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_SAR, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_SHR, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_ADD, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_SUB, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_MUL, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_DIV, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN_MOD, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)

	return p
}

func registerPrefixElementaryTypes(p *parser) {
	for _, tkType := range token.GetElementaryTypes() {
		p.registerPrefix(tkType, p.parseElementaryTypeExpression)
	}
}

func (p *parser) ToggleTracing() {
	p.trace = !p.trace
}

func (p *parser) nextToken() {
	p.currTkn = p.peekTkn
	p.peekTkn = p.l.NextToken()
}

func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace("parseFile"))
	}

	file := &ast.File{}
	file.SourceFile = p.file
	file.Declarations = []ast.Declaration{}

	for p.currTkn.Type != token.EOF {
		decl := p.parseSourceUnitDeclaration()
		if decl != nil {
			file.Declarations = append(file.Declarations, decl)
		}
		p.nextToken()
	}

	return file
}

func (p *parser) parseSourceUnitDeclaration() ast.Declaration {
	// Cases below should match elements outlined in the
	// 'rule source-unit' in Solidity Grammar.
	// TODO: Implement remaining SourceUnit elements.
	switch tk := p.currTkn.Type; tk {
	default:
		// TODO: Once this function is fully implemented, the default case
		// should only be hit on errors. Add the parses error then.
		p.addError(p.currTkn.Pos, "Unhandled declaration type in the SourceUnit: "+p.currTkn.Literal)
		return nil

	case token.COMMENT_LITERAL:
		// TODO Parse comments; skip for now
		return nil

	case token.PRAGMA:
		// TODO parse pragma; skip for now
		for !p.currTknIs(token.SEMICOLON) {
			p.nextToken()
		}
		return nil

	// import-directive

	case token.USING: // TODO: finish using-directive
		return p.parseUsingForDirective()

	case token.CONTRACT, token.ABSTRACT: // contract-definition
		return p.parseContractDeclaration()

		// interface-definition
		// library-definition

	case token.FUNCTION: // function-definition
		return p.parseFunctionDeclaration()

		// constant-variable-declaration
		// struct-definition
		// enum-definition
		// user-defined-value-type-definition
		// error-definition
	case token.EVENT: // event-definition
		return p.parseEventDeclaration()
	}
}

func (p *parser) parseContractDeclaration() *ast.ContractDeclaration {
	if p.trace {
		defer un(trace("parseContractDeclaration"))
	}

	decl := &ast.ContractDeclaration{}
	base := ast.ContractBase{}

	// parser is sitting either on the 'Contract' or 'Abstract' keyword.

	// If it is currently 'Abstract', the next will be 'Contract'
	if p.peekTknIs(token.CONTRACT) {
		decl.Abstract = true
		base.Pos = p.currTkn.Pos
		p.nextToken() // move to 'Contract'
	} else {
		// Pos of 'Contract'
		base.Pos = p.currTkn.Pos
	}

	// Parses is on the 'Contract' in both cases now.

	if !p.expectPeek(token.IDENTIFIER) {
		// TODO: See if this line must be kept. Most likely not.
		p.addError(p.currTkn.Pos, "Expected an identifier after Contract keyword. Got: "+p.peekTkn.Literal)
		// Move to get out of error state
		p.nextToken()
	}

	base.Name = &ast.Identifier{
		Pos:   p.currTkn.Pos,
		Value: p.currTkn.Literal,
	}

	// Parse inheritance, if any
	if p.peekTknIs(token.IS) {
		p.nextToken() // Move to 'IS'

		for {
			if !p.expectPeek(token.IDENTIFIER) {
				p.addError(p.currTkn.Pos, "Expected an identifier after IS keyword. Got: "+p.peekTkn.Literal)
				break
			}

			decl.Parents = append(decl.Parents, &ast.Identifier{
				Pos:   p.currTkn.Pos,
				Value: p.currTkn.Literal,
			})

			if !p.peekTknIs(token.COMMA) {
				break
			}

			p.nextToken() // Move past the comma
		}
	}

	// Parses either went through the ihneritance branch, or is still sitting on
	// the identifier.

	if p.expectPeek(token.LBRACE) {
		base.Body = p.parseContractBody()
	}

	decl.ContractBase = base

	return decl
}

func (p *parser) parseContractBody() *ast.ContractBody {
	if p.trace {
		defer un(trace("parseContractBody"))
	}

	// parser is sitting on the LBRACE
	body := &ast.ContractBody{
		LeftBrace: p.currTkn.Pos,
	}

	p.nextToken() // Move past LBRACE

	decls := []ast.Declaration{}

	for {
		// The cases below should mimic 1:1 elements outlined in:
		// 'rule contract-body-element' from Solidity Grammar page.
		// TODO: Implement remaining ContractBody elements.
		switch tk := p.currTkn.Type; {
		default:
			// TODO Once this function is fully implemented, throw parses errors
			// when default case is hit.
			p.addError(p.currTkn.Pos, "Unhandled declaration in contract's body: "+p.currTkn.Literal)
			p.nextToken()
		case tk == token.COMMENT_LITERAL:
			// TODO Parse comments
			p.nextToken()

		case tk == token.CONSTRUCTOR: // Constructor definition
			for !p.currTknIs(token.RBRACE) {
				p.nextToken()
			}
			p.nextToken() // Move past RBRACE

		case tk == token.FUNCTION: // Function definition
			decls = append(decls, p.parseFunctionDeclaration())
			p.nextToken() // Move past RBRACE

			// Modifier definition
			// fallback-function-definition
			// receive-function-definition
			// struct-definition
			// enum-definition
			// user-defined-value-type-definition

		case token.IsElementaryType(tk): // state-variable-declaration
			decls = append(decls, p.parseStateVariableDeclaration())
			p.nextToken() // Move past semicolon

		case tk == token.EVENT: // event-definition
			decls = append(decls, p.parseEventDeclaration())
			p.nextToken() // Move past semicolon

			// error-definition

		case tk == token.USING: // using-directive
			decls = append(decls, p.parseUsingForDirective())
			p.nextToken() // Move past semicolon

		case tk == token.RBRACE: // End of contract's body
			body.Declarations = decls
			body.RightBrace = p.currTkn.Pos

			return body
		}
	}
}

func (p *parser) parseUsingForDirective() *ast.UsingForDirective {
	if p.trace {
		defer un(trace("parseUsingForDirective"))
	}

	// The parser is sitting on the 'using' keyword.
	dir := &ast.UsingForDirective{
		Pos: p.currTkn.Pos,
	}

	p.nextToken() // Consume 'using'

	// The next token is either an identifier (for a library) or a '{' for an item list.
	if p.currTknIs(token.LBRACE) {
		dir.List = p.parseUsingForList()
	} else if p.currTknIs(token.IDENTIFIER) {
		dir.LibraryName = &ast.Identifier{Pos: p.currTkn.Pos, Value: p.currTkn.Literal}
		p.nextToken() // Consume library identifier
	} else {
		p.addError(p.currTkn.Pos, "expected library name or '{' after 'using'")
		return nil
	}

	if !p.currTknIs(token.FOR) {
		p.addError(p.currTkn.Pos, "expected 'for' after 'using' declaration")
		return nil
	}
	p.nextToken() // Consume 'for'

	// Expect either a type or the wildcard '*'.
	if p.currTknIs(token.MUL) {
		dir.IsWildcard = true
		p.nextToken() // Consume '*'
	} else {
		dir.ForType = p.parseTypeName()
	}

	// Check for optional 'global' keyword.
	if p.currTknIs(token.GLOBAL) {
		dir.IsGlobal = true
		p.nextToken() // Consume 'global'
	}

	if !p.currTknIs(token.SEMICOLON) {
		p.addError(p.currTkn.Pos, "expected ';' to terminate using directive")
		return nil
	}

	dir.Semicolon = p.currTkn.Pos

	return dir
}

func (p *parser) parseUsingForList() []*ast.UsingForObject {
	if p.trace {
		defer un(trace("parseUsingForList"))
	}

	var items []*ast.UsingForObject
	p.nextToken() // Consume '{'

	for !p.currTknIs(token.RBRACE) {
		item := &ast.UsingForObject{}
		if !p.currTknIs(token.IDENTIFIER) {
			p.addError(p.currTkn.Pos, "expected identifier in using list")
			return nil
		}
		item.Path = &ast.Identifier{Pos: p.currTkn.Pos, Value: p.currTkn.Literal}
		p.nextToken() // Consume identifier path

		if p.currTknIs(token.AS) {
			p.nextToken() // Consume 'as'
			if !token.IsUserDefinableOperator(p.currTkn.Type) {
				p.addError(p.currTkn.Pos, "expected a user-definable operator, got: "+p.currTkn.Literal)
				return nil
			}
			item.Alias = p.currTkn
			p.nextToken() // Consume alias
		}

		items = append(items, item)

		if p.currTknIs(token.COMMA) {
			p.nextToken() // Consume ','
		} else if !p.currTknIs(token.RBRACE) {
			p.addError(p.currTkn.Pos, "expected ',' or '}' in using list")
			return nil
		}
	}

	if !p.currTknIs(token.RBRACE) {
		p.addError(p.currTkn.Pos, "expected '}' to close using list")
		return nil
	}
	p.nextToken() // Consume '}'

	return items
}

func (p *parser) parseTypeName() ast.Type {
	if token.IsElementaryType(p.currTkn.Type) {
		t := &ast.ElementaryType{
			Pos:  p.currTkn.Pos,
			Kind: p.currTkn,
		}
		p.nextToken() // Consume the type token
		return t
	}

	if p.currTknIs(token.IDENTIFIER) {
		t := &ast.UserDefinedType{
			Name: &ast.Identifier{Pos: p.currTkn.Pos, Value: p.currTkn.Literal},
		}
		p.nextToken() // Consume the identifier token
		return t
	}

	p.addError(p.currTkn.Pos, "expected a type name (e.g., uint256 or MyStruct)")
	return nil
}

func (p *parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
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
		Pos:   p.currTkn.Pos,
		Value: p.currTkn.Literal,
	}

	// 3. ( Param List )
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	params := &ast.ParamList{}
	params.Opening = p.currTkn.Pos

	for !p.peekTknIs(token.RPAREN) {
		// One loop iteration parses one parameter.
		prm := &ast.Param{}
		prm.DataLocation = ast.NO_DATA_LOCATION // Explicit default value

		// param list looks like this:
		// ( typeName <<data location>> <<identifier>>, ... )
		// The typeName is required while data location and identifier are optional.

		// TODO: Param list parsing could be extracted to a separate
		// function since it is used in other places as well e.g. fallback,
		// receive functions, modifiers and return values.

		// TODO: Params can have other types than elementary types. For example:
		// - user defined like Contract names and structs
		// - function types
		// - arrays of other types
		// - mappings (?)
		if !token.IsElementaryType(p.peekTkn.Type) {
			p.addError(p.peekTkn.Pos, "Fn param: expected elementary type, got: "+p.peekTkn.Literal)
			return nil
		}
		p.nextToken() // Move to the type name.

		prm.Type = &ast.ElementaryType{
			Pos: p.currTkn.Pos,
			Kind: token.Token{
				Type:    p.currTkn.Type,
				Literal: p.currTkn.Literal,
				Pos:     p.currTkn.Pos,
			},
		}

		if token.IsDataLocation(p.peekTkn.Type) {
			p.nextToken()
			switch p.currTkn.Type {
			case token.STORAGE:
				prm.DataLocation = ast.Storage
			case token.MEMORY:
				prm.DataLocation = ast.Memory
			case token.CALLDATA:
				prm.DataLocation = ast.Calldata
			}
		}

		if p.peekTknIs(token.IDENTIFIER) {
			p.nextToken()
			prm.Name = &ast.Identifier{
				Pos:   p.currTkn.Pos,
				Value: p.currTkn.Literal,
			}
		}

		params.List = append(params.List, prm)

		if p.peekTknIs(token.COMMA) {
			p.nextToken()
		}
	}

	p.nextToken() // Move to the closing parenthesis.
	params.Closing = p.currTkn.Pos

	// 4. Visibility, State Mutability, Modifier Invocation, Override, Virtual

	// TODO: Parse stuff after params before body block.

	// 5. Returns ( Param List )

	// TODO: Parse returned results.

	// 6. Body block
	for !p.currTknIs(token.LBRACE) {
		p.nextToken()
	}

	fnBody := p.parseBlockStatement()

	decl.Params = params
	decl.Body = fnBody
	return decl
}

func (p *parser) parseStateVariableDeclaration() *ast.StateVariableDeclaration {
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
		Pos: p.currTkn.Pos,
		Kind: token.Token{
			Type:    p.currTkn.Type,
			Literal: p.currTkn.Literal,
			Pos:     p.currTkn.Pos,
		},
	}

	p.nextToken()

	// We might be sitting on the variable name OR the visibility specifier OR the mutability specifier

	// Visibility and mutability specifiers are flexible. Both are valid:
	// uint256 public constant x = 10;
	// uint256 constant public x = 10;

	for {
		switch tkType := p.currTkn.Type; {
		default:
			p.addError(p.currTkn.Pos, "Unexpected token: "+p.currTkn.Literal)
		case tkType == token.IDENTIFIER:
			decl.Name = &ast.Identifier{
				Pos:   p.currTkn.Pos,
				Value: p.currTkn.Literal,
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
			// TODO: Handle override. It requires changes in the AST.
			p.nextToken()
		case tkType == token.ASSIGN:
			p.nextToken()
			decl.Value = p.parseExpression(LOWEST)
			p.nextToken()
		case tkType == token.SEMICOLON:
			return decl
		}
	}
}

func (p *parser) parseEventDeclaration() *ast.EventDeclaration {
	if p.trace {
		defer un(trace("parseEventDeclaration"))
	}
	eventDecl := &ast.EventDeclaration{
		Pos: p.currTkn.Pos, // position of the "event" keyword
	}

	if !p.expectPeek(token.IDENTIFIER) { // if all good, move to identifier
		return nil
	}

	eventDecl.Name = &ast.Identifier{
		Pos:   p.currTkn.Pos,
		Value: p.currTkn.Literal,
	}

	if !p.expectPeek(token.LPAREN) { // if all good, move to opening parenthesis
		return nil
	}

	params := &ast.EventParamList{}
	params.Opening = p.currTkn.Pos

	for !p.peekTknIs(token.RPAREN) {
		eventParam := &ast.EventParam{}

		p.nextToken() // move past opening parenthesis
		if !token.IsElementaryType(p.currTkn.Type) {
			p.addError(p.currTkn.Pos, "Event param: expected elementary type after opening parenthesis, got: "+p.currTkn.Literal)
			return nil
		}

		eventParam.Type = &ast.ElementaryType{
			Pos: p.currTkn.Pos,
			Kind: token.Token{
				Type:    p.currTkn.Type,
				Literal: p.currTkn.Literal,
				Pos:     p.currTkn.Pos,
			},
		}

		// Indexed is an optional keyword
		if p.peekTknIs(token.INDEXED) {
			p.nextToken()
			eventParam.IsIndexed = true
		}

		// Param name is optional
		if p.peekTknIs(token.IDENTIFIER) {
			p.nextToken()
			eventParam.Name = &ast.Identifier{
				Pos:   p.currTkn.Pos,
				Value: p.currTkn.Literal,
			}
		}

		params.List = append(params.List, eventParam)

		if p.peekTknIs(token.COMMA) {
			p.nextToken() // Move to next param
		}
	}

	p.nextToken() // Move to closing parenthesis.
	params.Closing = p.currTkn.Pos

	eventDecl.Params = params

	if p.peekTknIs(token.ANONYMOUS) {
		p.nextToken() // Move to anonymous keyword
		eventDecl.IsAnonymous = true
	}

	if !p.expectPeek(token.SEMICOLON) { // if all good, move to semicolon
		return nil
	}

	return eventDecl
}

func (p *parser) parseStatement() ast.Statement {
	switch tkType := p.currTkn.Type; {
	default:
		return p.parseExpressionStatement()
	case token.IsElementaryType(tkType):
		// TODO: Implement other types that variables can have.
		// TODO: return address(0) and similar should be handled here
		return p.parseVariableDeclarationStatement()
	case tkType == token.LPAREN:
		return p.parseVariableDeclarationTupleStatement()
	case tkType == token.LBRACE:
		return p.parseBlockStatement()
	case tkType == token.RETURN:
		return p.parseReturnStatement()
	case tkType == token.IF:
		return p.parseIfStatement()
	case tkType == token.EMIT:
		return p.parseEmitStatement()
	}
}

func (p *parser) parseBlockStatement() *ast.BlockStatement {
	if p.trace {
		defer un(trace("parseBlockStatement"))
	}
	blockStmt := &ast.BlockStatement{}
	blockStmt.LeftBrace = p.currTkn.Pos

	p.nextToken()

	for {
		switch tkType := p.currTkn.Type; tkType {
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
		case token.COMMENT_LITERAL:
			p.nextToken()
			continue
		case token.UNCHECKED:
			// Move to the LBRACE.
			p.nextToken()
			stmt := p.parseUncheckedBlockStatement()
			if stmt != nil {
				blockStmt.Statements = append(blockStmt.Statements, stmt)
			}
			// Block parsing ends on RBRACE, so advance to the next token.
			p.nextToken()
		case token.RBRACE:
			// We have reached the end of the block.
			blockStmt.RightBrace = p.currTkn.Pos
			return blockStmt
		}
	}
}

// Almost the same as parseBlockStatement, but we don't allow nested unchecked
func (p *parser) parseUncheckedBlockStatement() *ast.UncheckedBlockStatement {
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
			p.addError(p.currTkn.Pos, "Nested unchecked blocks are not allowed.")
			p.nextToken() // Consume the offending token and continue parsing
			return nil
		case tkType == token.RBRACE:
			// We have reached the end of the block.
			blockStmt.RightBrace = p.currTkn.Pos
			return blockStmt
		}
	}
}

func (p *parser) parseVariableDeclarationStatement() *ast.VariableDeclarationStatement {
	if p.trace {
		defer un(trace("parseVariableDeclarationStatement"))
	}

	vdStmt := &ast.VariableDeclarationStatement{}
	vdStmt.DataLocation = ast.NO_DATA_LOCATION // assign default value

	vdStmt.Type = &ast.ElementaryType{
		Pos: p.currTkn.Pos,
		Kind: token.Token{
			Type:    p.currTkn.Type,
			Literal: p.currTkn.Literal,
			Pos:     p.currTkn.Pos,
		},
	}

	if token.IsDataLocation(p.peekTkn.Type) {
		p.nextToken()
		switch p.currTkn.Type {
		case token.STORAGE:
			vdStmt.DataLocation = ast.Storage
		case token.MEMORY:
			vdStmt.DataLocation = ast.Memory
		case token.CALLDATA:
			vdStmt.DataLocation = ast.Calldata
		}
	}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	vdStmt.Name = &ast.Identifier{
		Pos:   p.currTkn.Pos,
		Value: p.currTkn.Literal,
	}

	if p.peekTknIs(token.SEMICOLON) {
		p.nextToken()
		return vdStmt
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken() // Move past the ASSIGN token
	vdStmt.Value = p.parseExpression(LOWEST)

	if p.peekTknIs(token.SEMICOLON) {
		p.nextToken()
	}

	return vdStmt
}

func (p *parser) parseVariableDeclarationTupleStatement() *ast.VariableDeclarationTupleStatement {
	if p.trace {
		defer un(trace("parseVariableDeclarationTupleStatement"))
	}

	vdTupleStmt := &ast.VariableDeclarationTupleStatement{
		Opening: p.currTkn.Pos,
	}

	p.nextToken() // Consume '('

	if !p.currTknIs(token.RPAREN) {
		if token.IsElementaryType(p.currTkn.Type) {
			vdTupleStmt.Declarations = append(vdTupleStmt.Declarations, p.parseVariableDeclarationPart())
		} else {
			vdTupleStmt.Declarations = append(vdTupleStmt.Declarations, nil)
		}

		for p.currTknIs(token.COMMA) {
			p.nextToken() // Consume ','

			// Handle a trailing comma case, e.g., "(a,)"
			if p.currTknIs(token.RPAREN) {
				vdTupleStmt.Declarations = append(vdTupleStmt.Declarations, nil)
				break
			}

			// Parse the next element (or an empty slot for cases like ",,")
			if token.IsElementaryType(p.currTkn.Type) {
				vdTupleStmt.Declarations = append(vdTupleStmt.Declarations, p.parseVariableDeclarationPart())
			} else {
				vdTupleStmt.Declarations = append(vdTupleStmt.Declarations, nil)
			}
		}
	}

	// After the list, we must be at the closing parenthesis.
	if !p.currTknIs(token.RPAREN) {
		p.addError(p.currTkn.Pos, "expected ',' or ')' to close tuple declaration")
		return nil
	}
	vdTupleStmt.Closing = p.currTkn.Pos

	// Tuples must be declared with an initial value.
	if !p.expectPeek(token.ASSIGN) {
		p.addError(p.peekTkn.Pos, "variable declaration tuples must be initialized")
		return nil
	}

	p.nextToken() // Move past '='
	vdTupleStmt.Value = p.parseExpression(LOWEST)

	if p.peekTknIs(token.SEMICOLON) {
		p.nextToken()
	}

	return vdTupleStmt
}

// parseVariableDeclarationPart parses a variable declaration without the assignment
// or terminating semicolon. It's used for components within tuple declarations.
func (p *parser) parseVariableDeclarationPart() *ast.VariableDeclarationStatement {
	if p.trace {
		defer un(trace("parseVariableDeclarationPart"))
	}

	part := &ast.VariableDeclarationStatement{
		// Default to no location.
		DataLocation: ast.NO_DATA_LOCATION,
	}

	part.Type = &ast.ElementaryType{Pos: p.currTkn.Pos, Kind: p.currTkn}

	if token.IsDataLocation(p.peekTkn.Type) {
		p.nextToken()
		switch p.currTkn.Type {
		case token.STORAGE:
			part.DataLocation = ast.Storage
		case token.MEMORY:
			part.DataLocation = ast.Memory
		case token.CALLDATA:
			part.DataLocation = ast.Calldata
		}
	}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	part.Name = &ast.Identifier{Pos: p.currTkn.Pos, Value: p.currTkn.Literal}

	p.nextToken()

	return part
}

func (p *parser) parseReturnStatement() *ast.ReturnStatement {
	if p.trace {
		defer un(trace("parseReturnStatement()"))
	}

	retStmt := &ast.ReturnStatement{}
	retStmt.Pos = p.currTkn.Pos

	// Advance to the next token; parse the expression.
	p.nextToken()
	retStmt.Result = p.parseExpression(LOWEST)

	// Semicolon is optional on purpose to make it easier to type stuff into
	if p.peekTkn.Type == token.SEMICOLON {
		p.nextToken()
	}

	return retStmt
}

func (p *parser) parseExpressionStatement() *ast.ExpressionStatement {
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

func (p *parser) parseIfStatement() *ast.IfStatement {
	if p.trace {
		defer un(trace("parseIfStatement"))
	}

	// if (condition) { consequence } else { alternative }

	// 1. If keyword
	ifStmt := &ast.IfStatement{
		Pos: p.currTkn.Pos,
	}

	// 2. Parenthesis of condition should be next; consume the lparen token.
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// 3. Advance to condition; parse the condition.
	p.nextToken()
	ifStmt.Condition = p.parseExpression(LOWEST)

	// 4. Consume the rparen token.
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 5. Advance; parse the consequence statement.
	p.nextToken()
	ifStmt.Consequence = p.parseStatement()

	// 6. If there is an else, parse it.
	if p.peekTkn.Type == token.ELSE {
		p.nextToken() // Sitting on ELSE
		p.nextToken() // Sitting on the <<statement>> after ELSE
		ifStmt.Alternative = p.parseStatement()
	}

	return ifStmt
}

func (p *parser) parseEmitStatement() *ast.EmitStatement {
	if p.trace {
		defer un(trace("parseEmitStatement"))
	}

	// Emit expression is of the following format:
	// emit <<expression>> (call-argument-list) ;
	emitStmt := &ast.EmitStatement{
		Pos: p.currTkn.Pos, // parser is sitting on the emit keyword
	}

	p.nextToken() // Move past the emit keyword

	// Parse the expression that the parser is sitting on. It is an event name
	// along its expected call argument list. Parse expression will correctly
	// parse both.
	// TODO: Will this actually work correctly? Should the event work like
	// call expression?
	emitStmt.Expression = p.parseExpression(LOWEST)

	p.nextToken() // Move past semicolon

	return emitStmt
}

// expectPeek checks if the next token is of the expected type.
// If it is it advances the tokens.
func (p *parser) expectPeek(t token.TokenType) bool {
	if p.peekTkn.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be: %s, got: %s instead (at offset: %d)",
		t.String(), p.peekTkn.Type.String(), p.peekTkn.Pos)
	p.addError(p.peekTkn.Pos, msg)
}

// currTknIs checks if the current token is of the expected type.
func (p *parser) currTknIs(t token.TokenType) bool {
	return p.currTkn.Type == t
}

// peekTknIs checks if the next token is of the expected type.
func (p *parser) peekTknIs(t token.TokenType) bool {
	return p.peekTkn.Type == t
}

func (p *parser) registerPrefix(t token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

func (p *parser) registerInfix(t token.TokenType, fn infixParseFn) {
	p.infixParseFns[t] = fn
}

func (p *parser) Errors() ErrorList {
	return p.errors
}
