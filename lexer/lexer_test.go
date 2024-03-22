package lexer

import (
	"solparsor/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `
    Contract Vault {
        uint256 x;
        x = 5;
    }
    `
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		// Vault contract start
		{token.CONTRACT, "Contract"},
		{token.IDENTIFIER, "Vault"},
		{token.LBRACE, "{"},
		{token.UINT256, "uint256"},
		{token.IDENTIFIER, "x"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
		// Vault contract end
	}

	lexer := Lex(input)

	for i, tt := range tests {
		tkn := lexer.NextToken()

		if tkn.Type != tt.expectedType {
			t.Errorf("tests[%d] - token type wrong. expected: %s, got: %s", i, token.Tokens[tt.expectedType], token.Tokens[tkn.Type])
		}

		if tkn.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected: %s, got: %s", i, tt.expectedLiteral, tkn.Literal)
		}
	}
}
