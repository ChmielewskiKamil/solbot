package main

import (
	"fmt"
	// "os"
	"solparsor/lexer"
	// "solparsor/repl"
	"solparsor/token"
)

func main() {
	// repl.Start(os.Stdin)
	input := ``

	lexer := lexer.Lex(input)
	for {
		tkn := lexer.NextToken()
		fmt.Printf("Token: %s, at position: %d, with type: %s\n", tkn.String(), tkn.Pos, token.Tokens[tkn.Type])

		if tkn.Type == token.EOF || tkn.Type == token.ILLEGAL {
			break
		}
	}
}
