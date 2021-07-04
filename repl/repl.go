package repl

import (
	"bufio"
	"fmt"
	"io"
	"nala/evaluator"
	"nala/lexer"
	"nala/object"
	"nala/parser"
)

const PROMPT = "=> "

func Start(in io.Reader, out io.Writer) {
	// this wraps the input with a Buffer that we can Scan?
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		if line == ".q" {
			io.WriteString(out, "Arigatōgozaimashita!\n")
			break
		}
		// pl := lexer.New(line)
		l := lexer.New(line)
		p := parser.New(l)

		// for tok := pl.NextToken(); tok.Type != token.EOF; tok = pl.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
		}
		res := evaluator.Eval(program, env)

		if res != nil {
			io.WriteString(out, res.Inspect()+"\n")
		}
	}
}

const CAT_FACE = ` A_A
(-.-)
 |-|
/   \
|     |   __
|  || |  |  \__
\_||_/_/
`

func printParseErrors(out io.Writer, errs []string) {
	io.WriteString(out, CAT_FACE)
	io.WriteString(out, "Whoops! What an antagonized cat!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errs {
		io.WriteString(out, "\t"+msg+"\n")
	}

}
