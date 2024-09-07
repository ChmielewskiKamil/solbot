package repl

import (
	"bufio"
	"io"
	"solbot/parser"
	"solbot/token"
)

const PROMPT = ">> "
const FN_BOILERPLATE = "function repl() public { "

// @TODO: After introducing the file handle, the REPL does not work.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		print(PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		src := FN_BOILERPLATE + line + " }"
		p := parser.Parser{}
		handle := token.NewFile("repl.sol", src)
		p.Init(handle)

		file := p.ParseFile()
		// if len(p.Errors()) > 0 {
		// 	p.Errors().Print()
		// 	continue
		// }

		io.WriteString(out, file.String())
		io.WriteString(out, "\n")
	}
}
