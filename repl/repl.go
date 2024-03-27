package repl

import (
	"bufio"
	"fmt"
	"io"
	"solparsor/lexer"
	"solparsor/token"
)

const PROMPT = ">> "

func Start(in io.Reader) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		l := lexer.Lex(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			// % is the indicator of the start of a format specifier
			// + is used to present struct field names
			// v specifies the default format: show values
			fmt.Printf("%+v\n", tok)
		}
	}
}
