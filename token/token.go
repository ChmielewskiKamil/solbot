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

// Solidity tokens based on the [solc tokens](https://github.com/ethereum/solidity/blob/afda6984723fca99e82ebf34d0aec1804f1f3ce6/liblangutil/Token.h#L67),
// with a couple of small tweaks e.g. ERROR, EOS->EOF.
const (
	ERROR TokenType = iota
	EOF
	COMMENT

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
	keyword_beg
	CONTRACT
	FUNCTION
	IF
	ELSE
	keyword_end

	// Elementary Type Keywords
	elementary_type_beg
	INT
	UINT
	BYTES
	ADDRESS
	STRING
	BOOL
	elementary_type_end

	IDENTIFIER // x, y, foo, bar, etc. not a keyword, not a reserved word
)

// Tokens returns a human readable token name, given the int token type.
var Tokens = [...]string{
	ERROR:      "ERROR",
	EOF:        "EOF",
	COMMENT:    "COMMENT",
	IDENTIFIER: "IDENTIFIER",
	CONTRACT:   "Contract",
	UINT:       "uint",
}

var keywords map[string]TokenType
var elementaryTypes map[string]TokenType

func init() {
	keywords = make(map[string]TokenType, keyword_end-(keyword_beg+1))
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[Tokens[i]] = i
	}

	elementaryTypes = make(map[string]TokenType, elementary_type_end-(elementary_type_beg+1))
	for i := elementary_type_beg + 1; i < elementary_type_end; i++ {
		elementaryTypes[Tokens[i]] = i
	}
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	if tok, ok := elementaryTypes[ident]; ok {
		return tok
	}
	return IDENTIFIER
}
