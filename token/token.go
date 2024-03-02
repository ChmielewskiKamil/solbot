package token

import "fmt"

// Position is the offset to the beginning of a token, starting from 0
type Position int
type TokenType int

type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

func (tkn Token) String() string {
	switch tkn.Type {
	case EOF:
		return "EOF"
	case ERROR:
		return "Error: " + tkn.Literal
	}

	if len(tkn.Literal) > 10 {
		// a quoted string with max of 10 characters, followed by "..."
		// e.g. input "Hello, World!", output "Hello, Wor"...
		return fmt.Sprintf("%.10q...", tkn.Literal)
	}

	return fmt.Sprintf("%q", tkn.Literal)
}

const (
	// Special tokens
	_ TokenType = iota
	ERROR
	EOF
	COMMENT

	IDENTIFIER // x, y, foo, bar, etc.

	// Binary Operators
	ASSIGN   // =
	PLUS     // +
	MINUS    // -
	ASTERISK // *
	FSLASH   // /

	// Operators: Comparison
	LT     // <
	GT     // >
	EQ     // ==
	NOT_EQ // !=

	// Delimiters
	COMMA     // ,
	SEMICOLON // ;
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }

	// Keywords
	CONTRACT
	FUNCTION

	// Elementary Types
	UINT
	INT
	ADDRESS
	STRING
	BOOL
)
