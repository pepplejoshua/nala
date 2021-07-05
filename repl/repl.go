package repl

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"nala/evaluator"
	"nala/lexer"
	"nala/object"
	"nala/parser"
	"os"
)

const PROMPT = "=> "

const FUNCSPATH = "./nl/functions.nl"

// const FUNCSPATH = "./nl/test.nl"

func Start(in io.Reader, out io.Writer) {
	// this wraps the input with a Buffer that we can Scan?
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	// pre read some of handwritten functions from files
	readNalaFunctions(env, out)

	args := os.Args[1:]

	if len(args) == 1 {
		fName := args[0]
		pth := "./nl/" + fName + ".nl"
		// fmt.Println(pth)
		src := getFileContents(pth)
		// fmt.Println(src)

		// pl := lexer.New(src)
		l := lexer.New(src)
		p := parser.New(l)

		// for tok := pl.NextToken(); tok.Type != token.EOF; tok = pl.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }
		prog := p.ParseProgram()
		if hasErrors(p, out) {
			io.WriteString(out, fmt.Sprintf("Couldn't read Nala Functions Source from %q", pth))
			printParseErrors(out, p.Errors())
		}

		res := evaluator.Eval(prog, env)
		if res != nil {
			io.WriteString(out, res.Inspect()+"\n")
		} else {
			io.WriteString(out, "NIL\n")
		}
		return
	}

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

		if hasErrors(p, out) {
			printParseErrors(out, p.Errors())
		} else {
			res := evaluator.Eval(program, env)
			if res != nil {
				io.WriteString(out, res.Inspect()+"\n")
			} else {
				io.WriteString(out, "NIL\n")
			}
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

func readNalaFunctions(env *object.Environment, out io.Writer) {
	src := getFileContents(FUNCSPATH)
	// pl := lexer.New(src)
	l := lexer.New(src)
	p := parser.New(l)

	// for tok := pl.NextToken(); tok.Type != token.EOF; tok = pl.NextToken() {
	// fmt.Printf("%+v\n", tok)
	// }

	prog := p.ParseProgram()
	if hasErrors(p, out) {
		io.WriteString(out, fmt.Sprintf("Couldn't read Nala Functions Source from %q", FUNCSPATH))
		printParseErrors(out, p.Errors())
	}

	evaluator.Eval(prog, env)
	// if r != nil {
	io.WriteString(out, "Loaded Nala Functions Source!\n\n")
	// } else {
	// io.WriteString(out, "Couldn't load and evaluate source...\n\n")
	// }
}

func getFileContents(location string) string {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func hasErrors(p *parser.Parser, out io.Writer) bool {
	return len(p.Errors()) != 0
}
