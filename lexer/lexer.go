package lexer

import (
	"fmt"
	"solparsor/token"
	"unicode/utf8"
)

const (
	eof        = 0
	contract   = "Contract"
	leftBrace  = "{"
	rightBrace = "}"
)

// The state represents where we are in the input and what we expect to see next.
// An action defines what we are going to do in that state given the input.
// After you execute the action, you will be in a new state.
// Combining the state and the action together results in a state function.
// The stateFn represents the state of the lexer as a function that returns the next state.
// It is a recursive definition.
type stateFn func(*lexer) stateFn

// The `run` function lexes the input by executing state functions
// until the state is nil.
func (l *lexer) run() {
	for state := lexSourceUnit; state != nil; {
		state = state(l)
	}
	// The lexer is done, so we close the channel.
	// It tells the caller (probably the parser),
	// that no more tokens will be delivered.
	close(l.tokens)
}

// The lexer holds the state of the scanner.
type lexer struct {
	input  string           // The string being scanned.
	start  int              // Start position of this token.Token; in a big string, this is the start of the current token.
	pos    int              // Current position in the input.
	width  int              // Width of last rune read from input.
	state  stateFn          // The current state function.
	tokens chan token.Token // Channel of scanned token.
}

func Lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token.Token, 2), // Buffered channel
	}
	println("Lexing input: ", input)
	fmt.Printf("Input length: %d\n\n", len(input))
	// This starts the state machine.
	go l.run()

	return l
}

func (l *lexer) NextToken() token.Token {
	for {
		select {
		case tkn := <-l.tokens:
			return tkn
		}
	}
}

// The `emit` function passes an token.Token back to the client.
func (l *lexer) emit(typ token.TokenType) {
	// The value is a slice of the input.
	l.tokens <- token.Token{
		Type:    typ,
		Literal: l.input[l.start:l.pos],
		Pos:     token.Position(l.start),
	}
	// Move ahead in the input after sending it to the caller.
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token.Token{
		Type:    token.ERROR,
		Literal: fmt.Sprintf(format, args...),
		Pos:     token.Position(l.start),
	}
	return nil
}

func lexSourceUnit(l *lexer) stateFn {
	for {
		switch char := l.readChar(); {
		case char == eof:
			l.emit(token.EOF)
			return nil
		case isWhitespace(char):
			l.ignore()
		case isLetter(char):
			l.state = lexSourceUnit
			return lexIdentifier
		}
	}
}

func lexIdentifier(l *lexer) stateFn {
	for {
		switch char := l.readChar(); {
		case isLetter(char):
			// Do nothing, just keep reading.
		case isDigit(char):
			// Do nothing, just keep reading.
			// We entered here so we know that the first char is a letter.
			// We can have digits after letters in the identifiers.
		default:
			// We are sitting on something different than alphanumeric so just go back.
			l.backup()
			l.emit(token.LookupIdent(l.input[l.start:l.pos]))
			// Go back to the previous state.
			return l.state
		}
	}
}

// func lexFile(l *lexer) stateFn {
// 	for {
// 		if strings.HasPrefix(l.input[l.pos:], contract) {
// 			// If there was some text before the Contract declaration,
// 			// we just skip it for now
// 			if l.pos > l.start {
// 				l.ignore()
// 			}
// 			return lexContract
// 		}
//
// 		if l.readChar() == eof {
// 			break
// 		}
// 	}
//
// 	// If there was some text before the EOF token,
// 	// and we don't know how to handle it, we just skip it for now.
// 	if l.pos > l.start {
// 		l.ignore()
// 	}
// 	// Reached the end of the input file.
// 	l.emit(token.EOF)
// 	return nil
// }
//
// func lexContract(l *lexer) stateFn {
// 	l.pos += len(contract)
// 	l.emit(token.CONTRACT)
// 	return lexContractDeclaration
// }
//
// func lexContractDeclaration(l *lexer) stateFn {
// 	for {
// 		if strings.HasPrefix(l.input[l.pos:], leftBrace) {
// 			return lexLeftBrace
// 		}
//
// 		switch char := l.readChar(); {
// 		case isWhitespace(char):
// 			l.ignore()
// 		case char == eof:
// 			return l.errorf("Contract declaration not finished")
// 		case isLetter(char):
// 			l.state = lexContractDeclaration
// 			return lexIdentifier
// 		}
// 	}
// }
//
// func lexIdentifier(l *lexer) stateFn {
// 	for {
// 		switch char := l.readChar(); {
// 		case isLetter(char):
// 			// Do nothing, just keep reading.
// 		default:
// 			// We are sitting on something different than alphanumeric so just go back.
// 			l.backup()
// 			l.emit(token.IDENTIFIER)
// 			// Go back to the previous state.
// 			return l.state
// 		}
// 	}
// }
//
// func lexLeftBrace(l *lexer) stateFn {
// 	l.pos += len(leftBrace)
// 	l.emit(token.LBRACE)
// 	return lexInsideBraces
// }
//
// func lexInsideBraces(l *lexer) stateFn {
// 	if strings.HasPrefix(l.input[l.pos:], rightBrace) {
// 		return lexRightBrace
// 	}
//
// 	for {
// 		switch char := l.readChar(); {
// 		case isLetter(char):
// 			l.state = lexInsideBraces
// 			return lexIdentifier
// 		case isDigit(char):
// 			l.state = lexInsideBraces
// 			return lexNumber
// 		case char == ';':
// 			l.emit(token.SEMICOLON)
// 		case isWhitespace(char):
// 			l.ignore()
// 		}
// 	}
// }
//
// func lexRightBrace(l *lexer) stateFn {
// 	l.pos += len(rightBrace)
// 	l.emit(token.RBRACE)
// 	// @TODO: If next char is ), we could be inside function params
// 	return lexFile
// }
//
// func lexNumber(l *lexer) stateFn {
// 	for {
// 		switch char := l.readChar(); {
// 		case isDigit(char):
// 			// Do nothing, just keep reading.
// 		default:
// 			// We are sitting on something different than a digit so just go back.
// 			l.backup()
// 			l.emit(token.UINT)
// 			// Go back to the previous state.
// 			return l.state
// 		}
// 	}
// }

// readChar reads the next rune from the input, advances the position
// and returns the rune.
func (l *lexer) readChar() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width

	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.readChar()
	l.backup()
	return r
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
