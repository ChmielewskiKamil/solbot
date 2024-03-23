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
// A list of small differences:
// - EOS (End of Source) is renamed to EOF (End of File).
// - NUMBER is split into DECIMAL_NUMBER and HEX_NUMBER.
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
	ANONYMOUS // For events: does not store event signature as topic
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
	ether_subdenominations_beg
	SUB_WEI    // wei
	SUB_GWEI   // gwei
	SUB_ETHER  // ether
	SUB_SECOND // seconds
	SUB_MINUTE // minutes
	SUB_HOUR   // hours
	SUB_DAY    // days
	SUB_WEEK   // weeks
	SUB_YEAR   // years
	ether_subdenominations_end

	// Elementary Type Keywords
	elementary_type_beg
	INT
	INT_8
	INT_16
	INT_24
	INT_32
	INT_40
	INT_48
	INT_56
	INT_64
	INT_72
	INT_80
	INT_88
	INT_96
	INT_104
	INT_112
	INT_120
	INT_128
	INT_136
	INT_144
	INT_152
	INT_160
	INT_168
	INT_176
	INT_184
	INT_192
	INT_200
	INT_208
	INT_216
	INT_224
	INT_232
	INT_240
	INT_248
	INT_256

	UINT
	UINT_8
	UINT_16
	UINT_24
	UINT_32
	UINT_40
	UINT_48
	UINT_56
	UINT_64
	UINT_72
	UINT_80
	UINT_88
	UINT_96
	UINT_104
	UINT_112
	UINT_120
	UINT_128
	UINT_136
	UINT_144
	UINT_152
	UINT_160
	UINT_168
	UINT_176
	UINT_184
	UINT_192
	UINT_200
	UINT_208
	UINT_216
	UINT_224
	UINT_232
	UINT_240
	UINT_248
	UINT_256

	BYTES
	BYTES_1
	BYTES_2
	BYTES_3
	BYTES_4
	BYTES_5
	BYTES_6
	BYTES_7
	BYTES_8
	BYTES_9
	BYTES_10
	BYTES_11
	BYTES_12
	BYTES_13
	BYTES_14
	BYTES_15
	BYTES_16
	BYTES_17
	BYTES_18
	BYTES_19
	BYTES_20
	BYTES_21
	BYTES_22
	BYTES_23
	BYTES_24
	BYTES_25
	BYTES_26
	BYTES_27
	BYTES_28
	BYTES_29
	BYTES_30
	BYTES_31
	BYTES_32

	STRING
	ADDRESS
	BOOL
	FIXED
	UFIXED
	FIXED_MxN // @TODO: Handle fixed point numbers
	UFIXED_MxN
	elementary_type_end

	// Literals
	TRUE_LITERAL   // true
	FALSE_LITERAL  // false
	DECIMAL_NUMBER // This is different from solc, which has just NUMBER, for both hex and decimal
	HEX_NUMBER
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

	// Punctuators/Delimeters
	LPAREN:       "(",
	RPAREN:       ")",
	LBRACKET:     "[",
	RBRACKET:     "]",
	LBRACE:       "{",
	RBRACE:       "}",
	COLON:        ":",
	SEMICOLON:    ";",
	PERIOD:       ".",
	CONDITIONAL:  "?",
	DOUBLE_ARROW: "=>", // e.g. Solidity uses => for mapping
	RIGHT_ARROW:  "->",

	// Assignment Operators
	ASSIGN:         "=",
	ASSIGN_BIT_OR:  "|=",
	ASSIGN_BIT_XOR: "^=",
	ASSIGN_BIT_AND: "&=",
	ASSIGN_SHL:     "<<=",
	ASSIGN_SAR:     ">>=",
	ASSIGN_SHR:     ">>>=",
	ASSIGN_ADD:     "+=",
	ASSIGN_SUB:     "-=",
	ASSIGN_MUL:     "*=",
	ASSIGN_DIV:     "/=",
	ASSIGN_MOD:     "%=",

	// Binary Operators
	COMMA:   ",",
	OR:      "||",
	AND:     "&&",
	BIT_OR:  "|",
	BIT_XOR: "^",
	BIT_AND: "&",
	SHL:     "<<",
	SAR:     ">>",
	SHR:     ">>>",
	ADD:     "+",
	SUB:     "-",
	MUL:     "*",
	DIV:     "/",
	MOD:     "%",
	EXP:     "**",

	// Comparison Operators
	EQUAL:                 "==",
	NOT_EQUAL:             "!=",
	LESS_THAN:             "<",
	GREATER_THAN:          ">",
	LESS_THAN_OR_EQUAL:    "<=",
	GREATER_THAN_OR_EQUAL: ">=",

	// Unary Operators
	NOT:     "!",
	BIT_NOT: "~",
	INC:     "++",
	DEC:     "--",
	DELETE:  "delete",

	// Inline Assembly Operators
	ASSEMBLY_ASSIGN: ":=",

	// Keywords
	ABSTRACT:    "",
	ANONYMOUS:   "",
	AS:          "",
	ASSEMBLY:    "",
	BREAK:       "",
	CATCH:       "catch",
	CONSTANT:    "constant",
	CONSTRUCTOR: "",
	CONTINUE:    "",
	CONTRACT:    "Contract",
	DO:          "",
	ELSE:        "",
	ENUM:        "",
	EMIT:        "",
	EVENT:       "",
	EXTERNAL:    "",
	FALLBACK:    "",
	FOR:         "",
	FUNCTION:    "function",
	HEX:         "",
	IF:          "",
	INDEXED:     "",
	INTERFACE:   "",
	INTERNAL:    "",
	IMMUTABLE:   "",
	IMPORT:      "",
	IS:          "",
	LIBRARY:     "",
	MAPPING:     "mapping",
	MEMORY:      "",
	MODIFIER:    "",
	NEW:         "",
	OVERRIDE:    "",
	PAYABLE:     "",
	PUBLIC:      "public",
	PRAGMA:      "",
	PRIVATE:     "private",
	PURE:        "pure",
	RECEIVE:     "",
	RETURN:      "",
	RETURNS:     "",
	STORAGE:     "",
	CALLDATA:    "",
	STRUCT:      "",
	THROW:       "",
	TRY:         "",
	TYPE:        "",
	UNCHECKED:   "",
	USING:       "",
	VIEW:        "",
	VIRTUAL:     "",
	WHILE:       "",

	// Ether Subdenominations
	SUB_WEI:    "wei",
	SUB_GWEI:   "gwei",
	SUB_ETHER:  "ether",
	SUB_SECOND: "seconds",
	SUB_MINUTE: "minutes",
	SUB_HOUR:   "hours",
	SUB_DAY:    "days",
	SUB_WEEK:   "weeks",
	SUB_YEAR:   "years",

	// Elementary Type Keywords
	INT:     "int",
	INT_8:   "int8",
	INT_16:  "int16",
	INT_24:  "int24",
	INT_32:  "int32",
	INT_40:  "int40",
	INT_48:  "int48",
	INT_56:  "int56",
	INT_64:  "int64",
	INT_72:  "int72",
	INT_80:  "int80",
	INT_88:  "int88",
	INT_96:  "int96",
	INT_104: "int104",
	INT_112: "int112",
	INT_120: "int120",
	INT_128: "int128",
	INT_136: "int136",
	INT_144: "int144",
	INT_152: "int152",
	INT_160: "int160",
	INT_168: "int168",
	INT_176: "int176",
	INT_184: "int184",
	INT_192: "int192",
	INT_200: "int200",
	INT_208: "int208",
	INT_216: "int216",
	INT_224: "int224",
	INT_232: "int232",
	INT_240: "int240",
	INT_248: "int248",
	INT_256: "int256",

	UINT:     "uint",
	UINT_8:   "uint8",
	UINT_16:  "uint16",
	UINT_24:  "uint24",
	UINT_32:  "uint32",
	UINT_40:  "uint40",
	UINT_48:  "uint48",
	UINT_56:  "uint56",
	UINT_64:  "uint64",
	UINT_72:  "uint72",
	UINT_80:  "uint80",
	UINT_88:  "uint88",
	UINT_96:  "uint96",
	UINT_104: "uint104",
	UINT_112: "uint112",
	UINT_120: "uint120",
	UINT_128: "uint128",
	UINT_136: "uint136",
	UINT_144: "uint144",
	UINT_152: "uint152",
	UINT_160: "uint160",
	UINT_168: "uint168",
	UINT_176: "uint176",
	UINT_184: "uint184",
	UINT_192: "uint192",
	UINT_200: "uint200",
	UINT_208: "uint208",
	UINT_216: "uint216",
	UINT_224: "uint224",
	UINT_232: "uint232",
	UINT_240: "uint240",
	UINT_248: "uint248",
	UINT_256: "uint256",

	BYTES:    "bytes",
	BYTES_1:  "bytes1",
	BYTES_2:  "bytes2",
	BYTES_3:  "bytes3",
	BYTES_4:  "bytes4",
	BYTES_5:  "bytes5",
	BYTES_6:  "bytes6",
	BYTES_7:  "bytes7",
	BYTES_8:  "bytes8",
	BYTES_9:  "bytes9",
	BYTES_10: "bytes10",
	BYTES_11: "bytes11",
	BYTES_12: "bytes12",
	BYTES_13: "bytes13",
	BYTES_14: "bytes14",
	BYTES_15: "bytes15",
	BYTES_16: "bytes16",
	BYTES_17: "bytes17",
	BYTES_18: "bytes18",
	BYTES_19: "bytes19",
	BYTES_20: "bytes20",
	BYTES_21: "bytes21",
	BYTES_22: "bytes22",
	BYTES_23: "bytes23",
	BYTES_24: "bytes24",
	BYTES_25: "bytes25",
	BYTES_26: "bytes26",
	BYTES_27: "bytes27",
	BYTES_28: "bytes28",
	BYTES_29: "bytes29",
	BYTES_30: "bytes30",
	BYTES_31: "bytes31",
	BYTES_32: "bytes32",

	ADDRESS: "address",
	BOOL:    "bool",
	STRING:  "string",

	// FIXED: ?,
	// UFIXED: ?,
	// FIXED_MxN: ?,
	// UFIXED_MxN: ?,

	// Literals
	TRUE_LITERAL:           "true",
	FALSE_LITERAL:          "false",
	DECIMAL_NUMBER:         "",
	HEX_NUMBER:             "HEX_NUMBER",
	STRING_LITERAL:         "",
	UNICODE_STRING_LITERAL: "",
	HEX_STRING_LITERAL:     "",
	COMMENT_LITERAL:        "",

	// Identifiers, not keywords, not reserved words
	IDENTIFIER: "IDENTIFIER",

	// Keywords reserved for future use
	AFTER:        "",
	ALIAS:        "",
	APPLY:        "",
	AUTO:         "",
	BYTE:         "",
	CASE:         "",
	COPY_OF:      "",
	DEFAULT:      "",
	DEFINE:       "",
	FINAL:        "",
	IMPLEMENTS:   "",
	IN:           "",
	INLINE:       "",
	LET:          "",
	MACRO:        "",
	MATCH:        "",
	MUTABLE:      "",
	NULL_LITERAL: "",
	OF:           "",
	PARTIAL:      "",
	PROMISE:      "",
	REFERENCE:    "",
	RELOCATABLE:  "",
	SEALED:       "",
	SIZE_OF:      "",
	STATIC:       "",
	SUPPORTS:     "",
	SWITCH:       "",
	TYPE_DEF:     "",
	TYPE_OF:      "",
	VAR:          "",

	// Yul-specific tokens, but not keywords
	LEAVE: "leave",

	// Experimental Solidity specific keywords
	CLASS:         "",
	INSTANTIATION: "",
	INTEGER:       "",
	ITSELF:        "",
	STATIC_ASSERT: "",
	BUILTIN:       "",
	FOR_ALL:       "",
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
	// Keywords are longer than 1 character
	if len(ident) > 1 {
		if tok, ok := keywords[ident]; ok {
			return tok
		}
		if tok, ok := elementaryTypes[ident]; ok {
			return tok
		}
	}
	return IDENTIFIER
}
