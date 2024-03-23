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
	case ILLEGAL:
		return "Error: " + tkn.Literal
	}

	if len(tkn.Literal) > 10 {
		// a quoted string with max of 10 characters, followed by "..."
		// e.g. input "Hello, World!", output "Hello, Wor"...
		return fmt.Sprintf("%.10q...", tkn.Literal)
	}

	return fmt.Sprintf("%q", tkn.Literal)
}

// Solidity tokens based on the [solc tokens].
// Token names are exactly the same. The only difference is that they are formatted with screaming snake case.
// EOS (End of Source) is renamed to EOF (End of File).
// @TODO: Decide what to do with elementary types like UINT_M, INT_M, BYTES_M etc.
// Solc handles them in [a dynamic way]. We could do the same, but it would add a bit of complexity to the lexer.
// On the other hand these could be hardcoded into the list as well.
//
// [solc tokens]: https://github.com/ethereum/solidity/blob/afda6984723fca99e82ebf34d0aec1804f1f3ce6/liblangutil/Token.cpp#L183-L226
// [a dynamic way]: https://github.com/ethereum/solidity/blob/afda6984723fca99e82ebf34d0aec1804f1f3ce6/liblangutil/Token.cpp#L183-L226
const (
	ILLEGAL TokenType = iota
	EOF

	// Punctuators/Delimeters
	LPAREN       // (
	RPAREN       // )
	LBRACKET     // [
	RBRACKET     // ]
	LBRACE       // {
	RBRACE       // }
	COLON        // :
	SEMICOLON    // ;
	PERIOD       // .
	CONDITIONAL  // ?
	DOUBLE_ARROW // => e.g. Solidity uses => for mapping
	RIGHT_ARROW  // ->

	// Assignment Operators
	ASSIGN         // =
	ASSIGN_BIT_OR  // |=
	ASSIGN_BIT_XOR // ^=
	ASSIGN_BIT_AND // &=
	ASSIGN_SHL     // <<=
	ASSIGN_SAR     // >>=
	ASSIGN_SHR     // >>>=
	ASSIGN_ADD     // +=
	ASSIGN_SUB     // -=
	ASSIGN_MUL     // *=
	ASSIGN_DIV     // /=
	ASSIGN_MOD     // %=

	// Binary Operators
	COMMA   // ,
	OR      // ||
	AND     // &&
	BIT_OR  // |
	BIT_XOR // ^
	BIT_AND // &
	SHL     // <<
	SAR     // >>
	SHR     // >>>
	ADD     // +
	SUB     // -
	MUL     // *
	DIV     // /
	MOD     // %
	EXP     // **

	// Comparison Operators
	EQUAL                 // ==
	NOT_EQUAL             // !=
	LESS_THAN             // <
	GREATER_THAN          // >
	LESS_THAN_OR_EQUAL    // <=
	GREATER_THAN_OR_EQUAL // >=

	// Unary Operators
	NOT     // !
	BIT_NOT // ~
	INC     // ++
	DEC     // --
	DELETE  // delete

	// Inline Assembly Operators
	ASSEMBLY_ASSIGN // :=

	// Keywords
	keyword_beg
	ABSTRACT
	ANONYMOUS
	AS
	ASSEMBLY
	BREAK
	CATCH
	CONSTANT
	CONSTRUCTOR
	CONTINUE
	CONTRACT
	DO
	ELSE
	ENUM
	EMIT
	EVENT
	EXTERNAL
	FALLBACK
	FOR
	FUNCTION
	HEX
	IF
	INDEXED
	INTERFACE
	INTERNAL
	IMMUTABLE
	IMPORT
	IS
	LIBRARY
	MAPPING
	MEMORY
	MODIFIER
	NEW
	OVERRIDE
	PAYABLE
	PUBLIC
	PRAGMA
	PRIVATE
	PURE
	RECEIVE
	RETURN
	RETURNS
	STORAGE
	CALLDATA
	STRUCT
	THROW
	TRY
	TYPE
	UNCHECKED
	USING
	VIEW
	VIRTUAL
	WHILE
	keyword_end

	// Ether Subdenominations
	SUB_WEI    // wei
	SUB_GWEI   // gwei
	SUB_ETHER  // ether
	SUB_SECOND // seconds
	SUB_MINUTE // minutes
	SUB_HOUR   // hours
	SUB_DAY    // days
	SUB_WEEK   // weeks
	SUB_YEAR   // years

	// Elementary Type Keywords
	elementary_type_beg
	INT
	UINT
	BYTES
	STRING
	ADDRESS
	BOOL
	FIXED
	UFIXED
	INT_M   // 0 < M && M <= 256 && M % 8 == 0
	UINT_M  // 0 < M && M <= 256 && M % 8 == 0
	BYTES_M // 0 < M && M <= 32
	FIXED_MxN
	UFIXED_MxN
	elementary_type_end

	// Literals
	TRUE_LITERAL  // true
	FALSE_LITERAL // false
	NUMBER
	STRING_LITERAL
	UNICODE_STRING_LITERAL
	HEX_STRING_LITERAL
	COMMENT_LITERAL

	IDENTIFIER // x, y, foo, bar, etc. not a keyword, not a reserved word

	// Keywords reserved for future use
	AFTER
	ALIAS
	APPLY
	AUTO
	BYTE
	CASE
	COPY_OF
	DEFAULT
	DEFINE
	FINAL
	IMPLEMENTS
	IN
	INLINE
	LET
	MACRO
	MATCH
	MUTABLE
	NULL_LITERAL
	OF
	PARTIAL
	PROMISE
	REFERENCE
	RELOCATABLE
	SEALED
	SIZE_OF
	STATIC
	SUPPORTS
	SWITCH
	TYPE_DEF
	TYPE_OF
	VAR

	// Yul-specific tokens, but not keywords
	LEAVE // leave

	// Experimental Solidity specific keywords
	CLASS
	INSTANTIATION
	INTEGER
	ITSELF
	STATIC_ASSERT
	BUILTIN
	FOR_ALL
)

var Tokens = [...]string{
	// Special tokens
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	// Literals
	COMMENT_LITERAL: "COMMENT",

	// Assignment Operators

	// Binary Operators

	// Unary Operators

	// Inline Assembly Operators

	// Ether Subdenominations

	// Keywords
	CONTRACT: "Contract",
	UINT:     "uint",
	// UINT256:   "uint256",
	INT:       "int",
	ADDRESS:   "address",
	FUNCTION:  "function",
	CONSTANT:  "constant",
	IMMUTABLE: "immutable",

	// Elementary Type Keywords

	// Identifiers, not keywords, not reserved words
	IDENTIFIER: "IDENTIFIER",

	// Punctuators
	LBRACE:    "{",
	RBRACE:    "}",
	SEMICOLON: ";",
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
