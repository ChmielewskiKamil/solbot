package repl

import (
	"bufio"
	"fmt"
	"io"
	"solbot/lexer"
	"solbot/token"
)

const PROMPT = ">> "

// @TODO: After introducing the file handle, the REPL does not work.
func Start(in io.Reader) {
	scanner := bufio.NewScanner(in)

	for {
		print(PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		// line := scanner.Text()

		l := lexer.Lex(nil) // @TODO: Figure out how to use the file handle in REPL

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			// % is the indicator of the start of a format specifier
			// + is used to present struct field names
			// v specifies the default format: show values
			fmt.Printf("%+v\n", tok)
		}
	}
}
