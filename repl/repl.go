package repl

import (
	"bufio"
	"io"
	"solbot/evaluator"
	"solbot/parser"
	"solbot/token"
)

const PROMPT = ">> "
const FN_BOILERPLATE = "function repl() public { "
const ASCII_ART = `
|￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣|
|   You fucked up, try again...    |
|＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿|
              \ (•◡•) / 
               \     / 
                 ————
                 |   |
                 |_  |_`

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
		if len(p.Errors()) > 0 {
			io.WriteString(out, ASCII_ART)
			io.WriteString(out, "\n")
			for _, err := range p.Errors() {
				io.WriteString(out, "\t"+err.Msg+"\n")
			}
			continue
		}

		evaluated := evaluator.Eval(file)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}
