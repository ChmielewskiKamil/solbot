package lexer

import (
	"solparsor/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	// This does not have to be a 100% valid Solidity syntax.
	input := `
    contract Vault {
        uint256 x;
        x = 5;
        address owner = 0xDEADBEEF;
        mapping(address => uint256) balances;
        function deposit(uint256 amount) public {
            balances[msg.sender] += amount;
        }

        assembly {
            function f(x) -> y { /* implementation */ }
        }
    }

    // Just a comment
    library SafeMath {
        i != 0;
        i++;
        i--;
        
        123_456e-1_8 |= 0x1_5 ^= &= /= 1_2e1 %= ; 

        // comment
        a < b > c <= d >= e;
        a <<= b >>= c >>>= d >>> e << f >> g;
        a -> b;
        a -= b;
        /* this 
        * is 
        * a multi-line
        * comment
        */
        a == b ? -c : (a, b ** c);
    }

    uint256 y;
    address immutable attacker1337;

    address constant UniswapV3Factory = 0x1F98431c8aD98523631AE4a59f267346ea31F984;
    `
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		// Vault contract start
		{token.CONTRACT, "contract"},
		{token.IDENTIFIER, "Vault"},
		{token.LBRACE, "{"},
		{token.UINT_256, "uint256"},
		{token.IDENTIFIER, "x"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "x"},
		{token.ASSIGN, "="},
		{token.DECIMAL_NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.ADDRESS, "address"},
		{token.IDENTIFIER, "owner"},
		{token.ASSIGN, "="},
		{token.HEX_NUMBER, "0xDEADBEEF"},
		{token.SEMICOLON, ";"},
		{token.MAPPING, "mapping"},
		{token.LPAREN, "("},
		{token.ADDRESS, "address"},
		{token.DOUBLE_ARROW, "=>"},
		{token.UINT_256, "uint256"},
		{token.RPAREN, ")"},
		{token.IDENTIFIER, "balances"},
		{token.SEMICOLON, ";"},
		{token.FUNCTION, "function"},
		{token.IDENTIFIER, "deposit"},
		{token.LPAREN, "("},
		{token.UINT_256, "uint256"},
		{token.IDENTIFIER, "amount"},
		{token.RPAREN, ")"},
		{token.PUBLIC, "public"},
		{token.LBRACE, "{"},
		{token.IDENTIFIER, "balances"},
		{token.LBRACKET, "["},
		{token.IDENTIFIER, "msg"},
		{token.PERIOD, "."},
		{token.IDENTIFIER, "sender"},
		{token.RBRACKET, "]"},
		{token.ASSIGN_ADD, "+="},
		{token.IDENTIFIER, "amount"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ASSEMBLY, "assembly"},
		{token.LBRACE, "{"},
		{token.FUNCTION, "function"},
		{token.IDENTIFIER, "f"},
		{token.LPAREN, "("},
		{token.IDENTIFIER, "x"},
		{token.RPAREN, ")"},
		{token.RIGHT_ARROW, "->"},
		{token.IDENTIFIER, "y"},
		{token.LBRACE, "{"},
		{token.COMMENT_LITERAL, "/* implementation */"},
		{token.RBRACE, "}"},
		{token.RBRACE, "}"},
		{token.RBRACE, "}"},
		// Vault contract end

		// SafeMath library start
		{token.COMMENT_LITERAL, "// Just a comment"},
		{token.LIBRARY, "library"},
		{token.IDENTIFIER, "SafeMath"},
		{token.LBRACE, "{"},
		{token.IDENTIFIER, "i"},
		{token.NOT_EQUAL, "!="},
		{token.DECIMAL_NUMBER, "0"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "i"},
		{token.INC, "++"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "i"},
		{token.DEC, "--"},
		{token.SEMICOLON, ";"},
		{token.DECIMAL_NUMBER, "123_456e-1_8"},
		{token.ASSIGN_BIT_OR, "|="},
		{token.HEX_NUMBER, "0x1_5"},
		{token.ASSIGN_BIT_XOR, "^="},
		{token.ASSIGN_BIT_AND, "&="},
		{token.ASSIGN_DIV, "/="},
		{token.DECIMAL_NUMBER, "1_2e1"},
		{token.ASSIGN_MOD, "%="},
		{token.SEMICOLON, ";"},
		{token.COMMENT_LITERAL, "// comment"},
		{token.IDENTIFIER, "a"},
		{token.LESS_THAN, "<"},
		{token.IDENTIFIER, "b"},
		{token.GREATER_THAN, ">"},
		{token.IDENTIFIER, "c"},
		{token.LESS_THAN_OR_EQUAL, "<="},
		{token.IDENTIFIER, "d"},
		{token.GREATER_THAN_OR_EQUAL, ">="},
		{token.IDENTIFIER, "e"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "a"},
		{token.ASSIGN_SHL, "<<="},
		{token.IDENTIFIER, "b"},
		{token.ASSIGN_SAR, ">>="},
		{token.IDENTIFIER, "c"},
		{token.ASSIGN_SHR, ">>>="},
		{token.IDENTIFIER, "d"},
		{token.SHR, ">>>"},
		{token.IDENTIFIER, "e"},
		{token.SHL, "<<"},
		{token.IDENTIFIER, "f"},
		{token.SAR, ">>"},
		{token.IDENTIFIER, "g"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "a"},
		{token.RIGHT_ARROW, "->"},
		{token.IDENTIFIER, "b"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "a"},
		{token.ASSIGN_SUB, "-="},
		{token.IDENTIFIER, "b"},
		{token.SEMICOLON, ";"},
		{token.COMMENT_LITERAL, "/* this \n        * is \n        * a multi-line\n        * comment\n        */"},
		{token.IDENTIFIER, "a"},
		{token.EQUAL, "=="},
		{token.IDENTIFIER, "b"},
		{token.CONDITIONAL, "?"},
		{token.SUB, "-"},
		{token.IDENTIFIER, "c"},
		{token.COLON, ":"},
		{token.LPAREN, "("},
		{token.IDENTIFIER, "a"},
		{token.COMMA, ","},
		{token.IDENTIFIER, "b"},
		{token.EXP, "**"},
		{token.IDENTIFIER, "c"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		// SafeMath library end

		// Variables outside of contract
		{token.UINT_256, "uint256"},
		{token.IDENTIFIER, "y"},
		{token.SEMICOLON, ";"},
		{token.ADDRESS, "address"},
		{token.IMMUTABLE, "immutable"},
		{token.IDENTIFIER, "attacker1337"},
		{token.SEMICOLON, ";"},
		{token.ADDRESS, "address"},
		{token.CONSTANT, "constant"},
		{token.IDENTIFIER, "UniswapV3Factory"},
		{token.ASSIGN, "="},
		{token.HEX_NUMBER, "0x1F98431c8aD98523631AE4a59f267346ea31F984"},
		{token.SEMICOLON, ";"},

		{token.EOF, ""},
	}

	lexer := Lex(input)

	for i, tt := range tests {
		tkn := lexer.NextToken()

		if tkn.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected: %s (%d), got: %s",
				i, token.Tokens[tt.expectedType], tt.expectedType, token.Tokens[tkn.Type])
		}

		if tkn.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected: %s, got: %s",
				i, tt.expectedLiteral, tkn.Literal)
		}
	}
}
